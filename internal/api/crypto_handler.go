package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"multi-processing-backend/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/slog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type CryptoService interface {
	List(ctx context.Context, page, limit int) ([]core.Crypto, int64, error)
	Create(ctx context.Context, initial string, name string, current_value float64) (core.Crypto, error)


	GetLatestCryptos(ctx context.Context, limit int) ([]core.Crypto, error)
}

type CryptoHandler struct {
	service CryptoService
}

func NewCryptoHandler(service CryptoService) *CryptoHandler {
	return &CryptoHandler{service: service}
}

func RegisterCryptoRoutes(rg *gin.RouterGroup, h *CryptoHandler) {
	cryptos := rg.Group("")
	{
		cryptos.GET("", h.ListCryptos)
		cryptos.GET("/ws", h.HandleCryptoWS)
	}
}

func (h *CryptoHandler) ListCryptos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	cryptos, total, err := h.service.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, cryptos)
}

func (h *CryptoHandler) CreateCrypto(c *gin.Context) {
	var req struct {
		Initial      string  `json:"initial" binding:"required"`
		Name         string  `json:"name" binding:"required"`
		CurrentValue float64 `json:"current_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	crypto, err := h.service.Create(c.Request.Context(), req.Initial, req.Name, req.CurrentValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, crypto)
}

func (h *CryptoHandler) HandleCryptoWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Websocket upgrade faailed", "error", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				cancel()
				break
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cryptos, err := h.service.GetLatestCryptos(ctx, 12)
			if err != nil {
				slog.Error("failed to fetch latest cryptos for ws", "error", err)
				continue
			}

			data, err := json.Marshal(cryptos)
			if err != nil {
				slog.Error("failed to marshal cryptos", "error", err)
			}

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				slog.Error("websocket write failed", "error", err)
				return
			}
		}
	}
}
