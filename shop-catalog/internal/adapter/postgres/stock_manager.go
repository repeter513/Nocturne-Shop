// Package postgres implements atomic stock reserve, release, and confirm.
// Пакет postgres реализует атомарное резервирование, снятие и подтверждение остатков.
package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/repeter513/shop-catalog/internal/domain"
)

// StockManager performs transactional stock reservation operations.
// StockManager выполняет транзакционные операции резервирования остатков.
type StockManager struct {
	// db runs WithTx for atomic multi-statement operations.
	// db выполняет WithTx для атомарных многоstatement операций.
	db *DB
	// prod locks product rows and decrements physical stock on confirm.
	// prod блокирует строки товаров и уменьшает физический остаток при confirm.
	prod *ProductRepo
	// res reads/writes reservation rows and sums active reserved quantities.
	// res читает/пишет строки резервов и суммирует active зарезервированные количества.
	res *ReservationRepo
}

// NewStockManager creates a StockManager with product and reservation repos.
// NewStockManager создаёт StockManager с репозиториями товаров и резервов.
func NewStockManager(db *DB, prod *ProductRepo, res *ReservationRepo) *StockManager {
	return &StockManager{db: db, prod: prod, res: res}
}

// ReserveWithTransaction creates or merges a stock reservation for an order.
// ReserveWithTransaction создаёт или объединяет резерв остатков для заказа.
// Flow: FOR UPDATE existing by order_id → merge OR lock products + check available → INSERT.
// Flow: FOR UPDATE existing по order_id → merge ИЛИ lock products + check available → INSERT.
func (m *StockManager) ReserveWithTransaction(
	ctx context.Context,
	items map[int64]int32,
	orderID int64,
	ttlSeconds int32,
) (*domain.StockReservation, error) {
	reservation := &domain.StockReservation{
		OrderID:   orderID,
		Items:     items,
		Status:    domain.ReservationActive,
		ExpiresAt: time.Now().Add(reservationTTL(ttlSeconds)),
	}

	err := m.db.WithTx(ctx, func(tx pgx.Tx) error {
		// Step 1: idempotent path — if reservation exists for order_id, merge new items.
		// Шаг 1: идемпотентный путь — если резерв для order_id существует, объединить новые позиции.
		existing, err := m.res.scanOne(ctx, tx, `
			SELECT id, order_id, items, status, expires_at, created_at
			FROM stock_reservations WHERE order_id = $1 FOR UPDATE`, orderID)
		if err != nil {
			return err
		}
		if existing != nil {
			return m.mergeReservation(ctx, tx, existing, items, ttlSeconds, reservation)
		}

		// Step 2: new reservation — lock each product and verify available >= qty.
		// Шаг 2: новый резерв — блокировка каждого товара и проверка available >= qty.
		for productID, qty := range items {
			if qty <= 0 {
				return domain.ErrInsufficientStock
			}
			stock, err := m.prod.getStockForUpdate(ctx, tx, productID)
			if err != nil {
				return err
			}
			reserved, err := m.res.sumActiveReservedTx(ctx, tx, productID)
			if err != nil {
				return err
			}
			if stock-reserved < qty {
				return domain.ErrInsufficientStock
			}
		}

		// Step 3: insert reservation row; physical stock unchanged until confirm.
		// Шаг 3: вставка строки резерва; физический остаток не меняется до confirm.
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `
			INSERT INTO stock_reservations (order_id, items, status, expires_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at`,
			orderID, itemsJSON, reservation.Status, reservation.ExpiresAt,
		).Scan(&reservation.ID, &reservation.CreatedAt)
	})
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// mergeReservation adds items to an existing active reservation for the same order.
// mergeReservation добавляет позиции к существующему активному резерву того же заказа.
// Extends expires_at; re-validates available stock accounting for alreadyHeld in this reservation.
// Продлевает expires_at; повторно проверяет available с учётом alreadyHeld в этом резерве.
func (m *StockManager) mergeReservation(
	ctx context.Context,
	tx pgx.Tx,
	existing *domain.StockReservation,
	items map[int64]int32,
	ttlSeconds int32,
	out *domain.StockReservation,
) error {
	if existing.Status != domain.ReservationActive {
		return fmt.Errorf("order %d reservation is not active", existing.OrderID)
	}
	if existing.ExpiresAt.Before(time.Now()) {
		return domain.ErrReservationExpired
	}

	merged := make(map[int64]int32, len(existing.Items)+len(items))
	for productID, qty := range existing.Items {
		merged[productID] = qty
	}
	for productID, qty := range items {
		if qty <= 0 {
			return domain.ErrInsufficientStock
		}
		stock, err := m.prod.getStockForUpdate(ctx, tx, productID)
		if err != nil {
			return err
		}
		reserved, err := m.res.sumActiveReservedTx(ctx, tx, productID)
		if err != nil {
			return err
		}
		alreadyHeld := existing.Items[productID]
		if stock-reserved+alreadyHeld < alreadyHeld+qty {
			return domain.ErrInsufficientStock
		}
		merged[productID] += qty
	}

	itemsJSON, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(reservationTTL(ttlSeconds))
	if err := tx.QueryRow(ctx, `
		UPDATE stock_reservations SET items = $2, expires_at = $3
		WHERE id = $1
		RETURNING id, order_id, created_at`,
		existing.ID, itemsJSON, expiresAt,
	).Scan(&out.ID, &out.OrderID, &out.CreatedAt); err != nil {
		return err
	}
	out.Items = merged
	out.Status = domain.ReservationActive
	out.ExpiresAt = expiresAt
	return nil
}

// ReleaseWithTransaction removes items from a reservation or releases it entirely.
// ReleaseWithTransaction снимает позиции из резерва или освобождает его полностью.
// Empty items map releases all; status → released or partially_released.
// Пустой items map снимает всё; status → released или partially_released.
func (m *StockManager) ReleaseWithTransaction(ctx context.Context, reservationID int64, items map[int64]int32) error {
	return m.db.WithTx(ctx, func(tx pgx.Tx) error {
		reservation, err := m.res.scanOne(ctx, tx, `
			SELECT id, order_id, items, status, expires_at, created_at
			FROM stock_reservations WHERE id = $1 FOR UPDATE`, reservationID)
		if err != nil {
			return err
		}
		if reservation == nil {
			return domain.ErrReservationNotFound
		}
		if reservation.Status != domain.ReservationActive && reservation.Status != domain.ReservationPartiallyReleased {
			return domain.ErrReservationAlreadyReleased
		}

		releaseItems := items
		if len(releaseItems) == 0 {
			releaseItems = reservation.Items
		}

		for productID, qty := range releaseItems {
			current, ok := reservation.Items[productID]
			if !ok || current < qty {
				return domain.ErrInsufficientStock
			}
			reservation.Items[productID] = current - qty
			if reservation.Items[productID] == 0 {
				delete(reservation.Items, productID)
			}
		}

		itemsJSON, err := json.Marshal(reservation.Items)
		if err != nil {
			return err
		}

		status := domain.ReservationPartiallyReleased
		if len(reservation.Items) == 0 {
			status = domain.ReservationReleased
		}

		tag, err := tx.Exec(ctx, `UPDATE stock_reservations SET items = $2, status = $3 WHERE id = $1`,
			reservationID, itemsJSON, status)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrReservationNotFound
		}
		return nil
	})
}

// ConfirmWithTransaction deducts physical stock and marks the reservation confirmed.
// ConfirmWithTransaction списывает физический остаток и помечает резерв подтверждённым.
// Error paths: not found, wrong status, expired expires_at, insufficient stock on decrement.
// Пути ошибок: not found, неверный status, просроченный expires_at, insufficient stock при decrement.
func (m *StockManager) ConfirmWithTransaction(ctx context.Context, reservationID int64) error {
	return m.db.WithTx(ctx, func(tx pgx.Tx) error {
		reservation, err := m.res.scanOne(ctx, tx, `
			SELECT id, order_id, items, status, expires_at, created_at
			FROM stock_reservations WHERE id = $1 FOR UPDATE`, reservationID)
		if err != nil {
			return err
		}
		if reservation == nil {
			return domain.ErrReservationNotFound
		}
		if reservation.Status != domain.ReservationActive && reservation.Status != domain.ReservationPartiallyReleased {
			if reservation.Status == domain.ReservationExpired {
				return domain.ErrReservationExpired
			}
			return domain.ErrReservationAlreadyReleased
		}
		if reservation.ExpiresAt.Before(time.Now()) {
			return domain.ErrReservationExpired
		}

		for productID, qty := range reservation.Items {
			if err := m.prod.decrementStockTx(ctx, tx, productID, qty); err != nil {
				return err
			}
		}

		tag, err := tx.Exec(ctx, `UPDATE stock_reservations SET status = $2, items = '{}'::jsonb WHERE id = $1`,
			reservationID, domain.ReservationConfirmed)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrReservationNotFound
		}
		return nil
	})
}
