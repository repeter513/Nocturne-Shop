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

type Handler struct {
	clients *client.Clients
}

func NewHandler(clients *client.Clients) *Handler {
	return &Handler{clients: clients}
}

func (h *Handler) Register(mux *http.ServeMux, corsOrigins string) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.Handle("POST /api/v1/auth/register", cors(corsOrigins)(http.HandlerFunc(h.register)))
	mux.Handle("POST /api/v1/auth/login", cors(corsOrigins)(http.HandlerFunc(h.login)))
	mux.Handle("POST /api/v1/auth/refresh", cors(corsOrigins)(http.HandlerFunc(h.refresh)))
	mux.Handle("GET /api/v1/auth/me", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.me))))

	mux.Handle("GET /api/v1/products", cors(corsOrigins)(http.HandlerFunc(h.listProducts)))
	mux.Handle("GET /api/v1/products/{id}", cors(corsOrigins)(http.HandlerFunc(h.getProduct)))
	mux.Handle("GET /api/v1/products/{id}/stock", cors(corsOrigins)(http.HandlerFunc(h.getStock)))
	mux.Handle("GET /api/v1/categories", cors(corsOrigins)(http.HandlerFunc(h.listCategories)))

	mux.Handle("GET /api/v1/cart", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.getCart))))
	mux.Handle("POST /api/v1/cart/items", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.addToCart))))
	mux.Handle("PUT /api/v1/cart/items/{productId}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.updateCartItem))))
	mux.Handle("DELETE /api/v1/cart/items/{productId}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.removeFromCart))))
	mux.Handle("DELETE /api/v1/cart", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.clearCart))))

	mux.Handle("POST /api/v1/orders", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.createOrder))))
	mux.Handle("POST /api/v1/orders/{id}/pay", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.payOrder))))
	mux.Handle("POST /api/v1/orders/{id}/cancel", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.cancelOrder))))
	mux.Handle("GET /api/v1/orders", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.listOrders))))
	mux.Handle("GET /api/v1/orders/{id}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.getOrder))))

	mux.Handle("GET /api/v1/payments", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.listPayments))))
	mux.Handle("GET /api/v1/payments/{id}", cors(corsOrigins)(requireAuth(http.HandlerFunc(h.getPayment))))
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body registerReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
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

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body loginReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
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

type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
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

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token, _ := stringsCutBearer(auth)
	valid, err := h.clients.Auth.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{Token: token})
	if err != nil {
		writeError(w, err)
		return
	}
	if !valid.GetIsValid() {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}
	resp, err := h.clients.Auth.GetUserInfo(r.Context(), &authv1.GetUserInfoRequest{UserId: valid.GetUserId()})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    resp.GetUser().GetId(),
		"email": resp.GetUser().GetEmail(),
	})
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	page := int32(queryInt(r, "page", 1))
	pageSize := int32(queryInt(r, "pageSize", 10))
	categoryID := queryInt64(r, "categoryId")

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

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	resp, err := h.clients.Catalog.GetProduct(r.Context(), &catalogv1.GetProductRequest{ProductId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetProduct())
}

func (h *Handler) getStock(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	resp, err := h.clients.Catalog.GetStock(r.Context(), &catalogv1.GetStockRequest{ProductId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetStock())
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.clients.Catalog.ListCategories(r.Context(), &catalogv1.ListCategoriesRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Cart.GetCart(ctx, &cartv1.GetCartRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetCart())
}

type cartItemReq struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

func (h *Handler) addToCart(w http.ResponseWriter, r *http.Request) {
	var body cartItemReq
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
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

func (h *Handler) updateCartItem(w http.ResponseWriter, r *http.Request) {
	productID, err := pathInt64(r, "productId")
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		Quantity int32 `json:"quantity"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
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

func (h *Handler) removeFromCart(w http.ResponseWriter, r *http.Request) {
	productID, err := pathInt64(r, "productId")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Cart.RemoveFromCart(ctx, &cartv1.RemoveFromCartRequest{ProductId: productID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetCart())
}

func (h *Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	if _, err := h.clients.Cart.ClearCart(ctx, &cartv1.ClearCartRequest{}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Order.CreateOrder(ctx, &orderv1.CreateOrderRequest{})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp.GetOrder())
}

func (h *Handler) payOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Order.PayOrder(ctx, &orderv1.PayOrderRequest{OrderId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetOrder())
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Order.CancelOrder(ctx, &orderv1.CancelOrderRequest{OrderId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetOrder())
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	page := int32(queryInt(r, "page", 1))
	pageSize := int32(queryInt(r, "pageSize", 10))
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
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

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Order.GetOrder(ctx, &orderv1.GetOrderRequest{OrderId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetOrder())
}

func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	page := int32(queryInt(r, "page", 1))
	pageSize := int32(queryInt(r, "pageSize", 10))
	orderID := queryInt64(r, "orderId")
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
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

func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ctx := withAuth(r.Context(), r.Header.Get("Authorization"))
	resp, err := h.clients.Payment.GetPayment(ctx, &paymentv1.GetPaymentRequest{PaymentId: id})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.GetPayment())
}

func stringsCutBearer(auth string) (string, bool) {
	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		return "", false
	}
	return auth[len(prefix):], true
}
