// Package http implements REST handlers and middleware for the BFF.
// Пакет http реализует REST-обработчики и middleware для BFF.
package http

import (
	"net/http"

	"github.com/repeter513/shop-BFF/internal/client"
	authv1 "github.com/repeter513/shop-proto/gen/go/auth/v1"
	cartv1 "github.com/repeter513/shop-proto/gen/go/cart/v1"
	catalogv1 "github.com/repeter513/shop-proto/gen/go/catalog/v1"
	orderv1 "github.com/repeter513/shop-proto/gen/go/order/v1"
	paymentv1 "github.com/repeter513/shop-proto/gen/go/payment/v1"
)

// Handler proxies REST requests to backend gRPC services.
// Handler проксирует REST-запросы в backend gRPC-сервисы.
type Handler struct {
	// clients holds typed gRPC stubs for auth, catalog, cart, order, payment.
	// clients содержит типизированные gRPC-стабы auth, catalog, cart, order, payment.
	clients *client.Clients
}

// NewHandler constructs a Handler with the given gRPC clients.
// NewHandler создаёт Handler с переданными gRPC-клиентами.
func NewHandler(clients *client.Clients) *Handler {
	return &Handler{clients: clients}
}

// Register mounts all BFF routes on the given mux.
// Register монтирует все маршруты BFF на переданный mux.
func (h *Handler) Register(mux *http.ServeMux, corsOrigins string) {
	// GET /health — liveness probe, no gRPC call.
	// GET /health — liveness probe, без вызова gRPC.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// --- Auth (auth gRPC: AuthService) ---
	// POST /api/v1/auth/register → AuthService.RegisterUser
	mux.Handle("POST /api/v1/auth/register", cors(corsOrigins)(http.HandlerFunc(h.register)))
	// POST /api/v1/auth/login → AuthService.LoginUser
	mux.Handle("POST /api/v1/auth/login", cors(corsOrigins)(http.HandlerFunc(h.login)))
	// POST /api/v1/auth/refresh → AuthService.RefreshToken
	mux.Handle("POST /api/v1/auth/refresh", cors(corsOrigins)(http.HandlerFunc(h.refresh)))
	// GET /api/v1/auth/me → ValidateToken + GetUserInfo (requireAuth)
	mux.Handle("GET /api/v1/auth/me", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.me))))

	// --- Catalog (catalog gRPC: CatalogService) — public, no JWT ---
	// GET /api/v1/products → CatalogService.ListProducts
	mux.Handle("GET /api/v1/products", cors(corsOrigins)(http.HandlerFunc(h.listProducts)))
	// GET /api/v1/products/{id} → CatalogService.GetProduct
	mux.Handle("GET /api/v1/products/{id}", cors(corsOrigins)(http.HandlerFunc(h.getProduct)))
	// GET /api/v1/products/{id}/stock → CatalogService.GetStock
	mux.Handle("GET /api/v1/products/{id}/stock", cors(corsOrigins)(http.HandlerFunc(h.getStock)))
	// GET /api/v1/categories → CatalogService.ListCategories
	mux.Handle("GET /api/v1/categories", cors(corsOrigins)(http.HandlerFunc(h.listCategories)))

	// --- Cart (cart gRPC: CartService) — JWT required, token forwarded via withAuth ---
	// GET /api/v1/cart → CartService.GetCart
	mux.Handle("GET /api/v1/cart", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.getCart))))
	// POST /api/v1/cart/items → CartService.AddToCart
	mux.Handle("POST /api/v1/cart/items", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.addToCart))))
	// PUT /api/v1/cart/items/{productId} → CartService.UpdateCartItem
	mux.Handle("PUT /api/v1/cart/items/{productId}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.updateCartItem))))
	// DELETE /api/v1/cart/items/{productId} → CartService.RemoveFromCart
	mux.Handle("DELETE /api/v1/cart/items/{productId}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.removeFromCart))))
	// DELETE /api/v1/cart → CartService.ClearCart
	mux.Handle("DELETE /api/v1/cart", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.clearCart))))

	// --- Orders (order gRPC: OrderService) — JWT required ---
	// POST /api/v1/orders → OrderService.CreateOrder
	mux.Handle("POST /api/v1/orders", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.createOrder))))
	// POST /api/v1/orders/{id}/pay → OrderService.PayOrder
	mux.Handle("POST /api/v1/orders/{id}/pay", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.payOrder))))
	// POST /api/v1/orders/{id}/cancel → OrderService.CancelOrder
	mux.Handle("POST /api/v1/orders/{id}/cancel", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.cancelOrder))))
	// GET /api/v1/orders → OrderService.ListOrders
	mux.Handle("GET /api/v1/orders", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.listOrders))))
	// GET /api/v1/orders/{id} → OrderService.GetOrder
	mux.Handle("GET /api/v1/orders/{id}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.getOrder))))

	// --- Payments (payment gRPC: PaymentService) — JWT required ---
	// GET /api/v1/payments → PaymentService.ListPayments
	mux.Handle("GET /api/v1/payments", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.listPayments))))
	// GET /api/v1/payments/{id} → PaymentService.GetPayment
	mux.Handle("GET /api/v1/payments/{id}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.getPayment))))
}

// registerReq is the JSON body for user registration.
// registerReq — JSON-тело для регистрации пользователя.
type registerReq struct {
	// Email is the user's login identifier.
	// Email — идентификатор входа пользователя.
	Email string `json:"email"`
	// Password is the plaintext credential (hashed server-side).
	// Password — пароль в открытом виде (хешируется на сервере).
	Password string `json:"password"`
}

// register creates a new user account.
// register создаёт новую учётную запись пользователя.
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body registerReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	// gRPC: AuthService.RegisterUser
	resp, err := h.clients.Auth.RegisterUser(r.Context(), &authv1.RegisterUserRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"userId": resp.GetUserId()})
}

// loginReq is the JSON body for user login.
// loginReq — JSON-тело для входа пользователя.
type loginReq struct {
	// Email is the registered email address.
	// Email — зарегистрированный адрес электронной почты.
	Email string `json:"email"`
	// Password is the user's password.
	// Password — пароль пользователя.
	Password string `json:"password"`
}

// login authenticates a user and returns tokens.
// login аутентифицирует пользователя и возвращает токены.
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body loginReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	// gRPC: AuthService.LoginUser → JWT access + refresh tokens
	resp, err := h.clients.Auth.LoginUser(r.Context(), &authv1.LoginUserRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken":  resp.GetAccessToken(),
		"refreshToken": resp.GetRefreshToken(),
		"expiresIn":    resp.GetExpiresIn(),
		"userId":       resp.GetUserId(),
	})
}

// refreshReq is the JSON body for token refresh.
// refreshReq — JSON-тело для обновления токена.
type refreshReq struct {
	// RefreshToken is the long-lived token issued at login.
	// RefreshToken — долгоживущий токен, выданный при входе.
	RefreshToken string `json:"refreshToken"`
}

// refresh exchanges a refresh token for new access credentials.
// refresh обменивает refresh-токен на новые учётные данные доступа.
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	// gRPC: AuthService.RefreshToken
	resp, err := h.clients.Auth.RefreshToken(r.Context(), &authv1.RefreshTokenRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken":  resp.GetAccessToken(),
		"refreshToken": resp.GetRefreshToken(),
		"expiresIn":    resp.GetExpiresIn(),
	})
}

// me returns the authenticated user's profile.
// me возвращает профиль аутентифицированного пользователя.
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token, _ := stringsCutBearer(auth)
	// gRPC: AuthService.ValidateToken — full JWT signature/expiry check
	valid, err := h.clients.Auth.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{Token: token})
	if err != nil {
		writeError(w, err)
		return
	}
	if !valid.GetIsValid() {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}
	// gRPC: AuthService.GetUserInfo — requires auth metadata
	resp, err := h.clients.Auth.GetUserInfo(withAuth(r.Context(), auth), &authv1.GetUserInfoRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    resp.GetUser().GetId(),
		"email": resp.GetUser().GetEmail(),
	})
}

// listProducts returns a paginated product catalog.
// listProducts возвращает постраничный каталог товаров.
func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	page := int32(queryInt(r, "page", 1))
	pageSize := int32(queryInt(r, "pageSize", 10))
	categoryID := queryInt64(r, "categoryId")

	// gRPC: CatalogService.ListProducts
	resp, err := h.clients.Catalog.ListProducts(r.Context(), &catalogv1.ListProductsRequest{
		Page:       page,
		PageSize:   pageSize,
		CategoryId: categoryID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// getProduct returns a single product by ID.
// getProduct возвращает один товар по ID.
func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	// gRPC: CatalogService.GetProduct
	resp, err := h.clients.Catalog.GetProduct(r.Context(), &catalogv1.GetProductRequest{ProductId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetProduct())
}

// getStock returns available stock for a product.
// getStock возвращает доступный остаток товара.
func (h *Handler) getStock(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	// gRPC: CatalogService.GetStock
	resp, err := h.clients.Catalog.GetStock(r.Context(), &catalogv1.GetStockRequest{ProductId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetStock())
}

// listCategories returns all product categories.
// listCategories возвращает все категории товаров.
func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	// gRPC: CatalogService.ListCategories
	resp, err := h.clients.Catalog.ListCategories(r.Context(), &catalogv1.ListCategoriesRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// getCart returns the authenticated user's cart.
// getCart возвращает корзину аутентифицированного пользователя.
func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: CartService.GetCart
	resp, err := h.clients.Cart.GetCart(ctx, &cartv1.GetCartRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetCart())
}

// cartItemReq is the JSON body for adding an item to the cart.
// cartItemReq — JSON-тело для добавления позиции в корзину.
type cartItemReq struct {
	// ProductID is the catalog product to add.
	// ProductID — товар каталога для добавления.
	ProductID int64 `json:"productId"`
	// Quantity is how many units to add.
	// Quantity — количество единиц для добавления.
	Quantity int32 `json:"quantity"`
}

// addToCart adds a product to the user's cart.
// addToCart добавляет товар в корзину пользователя.
func (h *Handler) addToCart(w http.ResponseWriter, r *http.Request) {
	var body cartItemReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: CartService.AddToCart
	resp, err := h.clients.Cart.AddToCart(ctx, &cartv1.AddToCartRequest{
		ProductId: body.ProductID,
		Quantity:  body.Quantity,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetCart())
}

// updateCartItem changes the quantity of a cart line item.
// updateCartItem изменяет количество позиции в корзине.
func (h *Handler) updateCartItem(w http.ResponseWriter, r *http.Request) {
	productID, err := pathInt64(r, "productId")
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		// Quantity is the new item count (0 removes the line).
		// Quantity — новое количество (0 удаляет позицию).
		Quantity int32 `json:"quantity"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: CartService.UpdateCartItem
	resp, err := h.clients.Cart.UpdateCartItem(ctx, &cartv1.UpdateCartItemRequest{
		ProductId: productID,
		Quantity:  body.Quantity,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetCart())
}

// removeFromCart removes a product from the user's cart.
// removeFromCart удаляет товар из корзины пользователя.
func (h *Handler) removeFromCart(w http.ResponseWriter, r *http.Request) {
	productID, err := pathInt64(r, "productId")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: CartService.RemoveFromCart
	resp, err := h.clients.Cart.RemoveFromCart(ctx, &cartv1.RemoveFromCartRequest{ProductId: productID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetCart())
}

// clearCart removes all items from the user's cart.
// clearCart удаляет все позиции из корзины пользователя.
func (h *Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: CartService.ClearCart
	if _, err := h.clients.Cart.ClearCart(ctx, &cartv1.ClearCartRequest{}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// createOrder creates a new order from the user's cart.
// createOrder создаёт новый заказ из корзины пользователя.
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: OrderService.CreateOrder — reserves stock for PAYMENT_RESERVE window
	resp, err := h.clients.Order.CreateOrder(ctx, &orderv1.CreateOrderRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp.GetOrder())
}

// payOrder processes payment for an order.
// payOrder обрабатывает оплату заказа.
func (h *Handler) payOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: OrderService.PayOrder → triggers payment microservice
	resp, err := h.clients.Order.PayOrder(ctx, &orderv1.PayOrderRequest{OrderId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetOrder())
}

// cancelOrder cancels a pending order.
// cancelOrder отменяет заказ в статусе pending.
func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: OrderService.CancelOrder — releases stock reservation
	resp, err := h.clients.Order.CancelOrder(ctx, &orderv1.CancelOrderRequest{OrderId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetOrder())
}

// listOrders returns paginated orders for the authenticated user.
// listOrders возвращает постраничный список заказов аутентифицированного пользователя.
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	page := int32(queryInt(r, "page", 1))
	pageSize := int32(queryInt(r, "pageSize", 10))
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: OrderService.ListOrders
	resp, err := h.clients.Order.ListOrders(ctx, &orderv1.ListOrdersRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// getOrder returns a single order by ID.
// getOrder возвращает один заказ по ID.
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: OrderService.GetOrder
	resp, err := h.clients.Order.GetOrder(ctx, &orderv1.GetOrderRequest{OrderId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetOrder())
}

// listPayments returns paginated payments, optionally filtered by order.
// listPayments возвращает постраничный список платежей с фильтром по заказу.
func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	page := int32(queryInt(r, "page", 1))
	pageSize := int32(queryInt(r, "pageSize", 10))
	orderID := queryInt64(r, "orderId")
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: PaymentService.ListPayments
	resp, err := h.clients.Payment.ListPayments(ctx, &paymentv1.ListPaymentsRequest{
		OrderId:  orderID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// getPayment returns a single payment by ID.
// getPayment возвращает один платёж по ID.
func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	// gRPC: PaymentService.GetPayment
	resp, err := h.clients.Payment.GetPayment(ctx, &paymentv1.GetPaymentRequest{PaymentId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetPayment())
}

// stringsCutBearer extracts the token from a Bearer Authorization header.
// stringsCutBearer извлекает токен из заголовка Authorization Bearer.
func stringsCutBearer(auth string) (string, bool) {
	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		return "", false
	}
	return auth[len(prefix):], true
}
