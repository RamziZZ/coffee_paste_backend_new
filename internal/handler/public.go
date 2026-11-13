package handler

import (
	"math"
	"net/http"
	"strconv"

	"coffeepaste/internal/format"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

type PublicHandler struct {
	menu    *store.MenuStore
	reviews *store.ReviewStore
}

func NewPublicHandler(menu *store.MenuStore, reviews *store.ReviewStore) *PublicHandler {
	return &PublicHandler{menu: menu, reviews: reviews}
}

// Slug dan judul seksi kategori dibuat di sini karena belum ada kolomnya di database.
// Kunci = menu_categories.name.
var categoryMeta = map[string]struct{ Slug, Title string }{
	"Coffee":     {"minuman-kopi", "Minuman Kopi"},
	"Non-Coffee": {"minuman-non-kopi", "Minuman Non Kopi"},
	"Makanan":    {"makanan", "Makanan"},
	"Pastry":     {"pastry-snack", "Pastry & Snack"},
}

type featuredItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       string  `json:"price"`
	Image       *string `json:"image"`
}

type categoryItem struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Price      string  `json:"price"`                // harga Regular
	PriceLarge string  `json:"priceLarge,omitempty"` // harga Large; kosong untuk makanan/pastry
	Image      *string `json:"image"`
}

type categoryResponse struct {
	ID     string         `json:"id"`
	Filter string         `json:"filter"`
	Icon   string         `json:"icon"`
	Title  string         `json:"title"`
	Items  []categoryItem `json:"items"`
}

type reviewResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	Product   string `json:"product"` // nama produk yang diulas; kosong untuk ulasan umum
	CreatedAt string `json:"createdAt"`
}

type reviewSummary struct {
	Average float64 `json:"average"` // 1 desimal
	Total   int     `json:"total"`
}

type reviewsPayload struct {
	Summary reviewSummary    `json:"summary"`
	Items   []reviewResponse `json:"items"`
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

// Featured: GET /api/menu/featured
func (h *PublicHandler) Featured(w http.ResponseWriter, r *http.Request) {
	const limit = 4

	items, err := h.menu.BestSellers(r.Context(), limit)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if len(items) == 0 { // belum ada penjualan: pakai is_featured
		items, err = h.menu.FeaturedFlagged(r.Context(), limit)
		if err != nil {
			httpx.ServerError(w, err)
			return
		}
	}

	out := make([]featuredItem, 0, len(items))
	for _, it := range items {
		out = append(out, featuredItem{
			ID:          it.ID,
			Name:        it.Name,
			Description: it.Description,
			Price:       format.Rupiah(it.Price),
			Image:       it.Image,
		})
	}
	httpx.JSON(w, http.StatusOK, out)
}

// Categories: GET /api/menu/categories
func (h *PublicHandler) Categories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.menu.Categories(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	out := make([]categoryResponse, 0, len(cats))
	for _, c := range cats {
		meta, ok := categoryMeta[c.Name]
		if !ok {
			meta.Slug = format.Slug(c.Name)
			meta.Title = c.Name
		}
		filter := c.FilterLabel
		if filter == "" {
			filter = c.Name
		}
		icon := c.Icon
		if icon == "" {
			icon = "coffee-cup"
		}

		items := make([]categoryItem, 0, len(c.Items))
		for _, it := range c.Items {
			ci := categoryItem{
				ID:    it.ID,
				Name:  it.Name,
				Price: format.Rupiah(it.Price),
				Image: it.Image,
			}
			if it.PriceLarge != nil {
				ci.PriceLarge = format.Rupiah(*it.PriceLarge)
			}
			items = append(items, ci)
		}
		out = append(out, categoryResponse{
			ID:     meta.Slug,
			Filter: filter,
			Icon:   icon,
			Title:  meta.Title,
			Items:  items,
		})
	}
	httpx.JSON(w, http.StatusOK, out)
}

// Filters: GET /api/menu/filters
func (h *PublicHandler) Filters(w http.ResponseWriter, r *http.Request) {
	cats, err := h.menu.Categories(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	out := []string{"Semua Menu"}
	for _, c := range cats {
		label := c.FilterLabel
		if label == "" {
			label = c.Name
		}
		out = append(out, label)
	}
	httpx.JSON(w, http.StatusOK, out)
}

// Reviews: GET /api/reviews?limit=6
// Bentuk: { summary: { average, total }, items: [{ id, name, rating, comment, product, createdAt }] }
// items: ulasan terbaru yang dipublikasikan, berkomentar, dan berbintang 4 ke atas (landing page).
// summary: dihitung dari SEMUA ulasan yang dipublikasikan, bukan hanya items.
func (h *PublicHandler) Reviews(w http.ResponseWriter, r *http.Request) {
	limit := 12
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = min(v, 50)
	}

	list, err := h.reviews.Published(r.Context(), 4, limit)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	sum, err := h.reviews.Summary(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	items := make([]reviewResponse, 0, len(list))
	for _, rv := range list {
		items = append(items, reviewResponse{
			ID:        rv.ID,
			Name:      rv.Name,
			Rating:    rv.Rating,
			Comment:   rv.Comment,
			Product:   rv.Product,
			CreatedAt: rv.CreatedAt.Format(format.ISO),
		})
	}
	httpx.JSON(w, http.StatusOK, reviewsPayload{
		Summary: reviewSummary{Average: round1(sum.Average), Total: sum.Total},
		Items:   items,
	})
}

type menuResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	CategoryID  string  `json:"categoryId"` // sama dengan "id" pada /menu/categories
	Price       int64   `json:"price"`      // harga Regular
	PriceLarge  *int64  `json:"priceLarge"` // harga Large; null untuk makanan/pastry
	Status      string  `json:"status"`     // tersedia | habis
	Image       *string `json:"image"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`      // rata-rata bintang (1 desimal); 0 bila belum ada ulasan
	RatingCount int     `json:"ratingCount"` // jumlah ulasan yang dipublikasikan
}

// categorySlug memberi id kategori yang sama dengan GET /api/menu/categories.
func categorySlug(name string) string {
	if meta, ok := categoryMeta[name]; ok {
		return meta.Slug
	}
	return format.Slug(name)
}

// Menu: GET /api/menu
// Daftar seluruh menu (termasuk yang habis) untuk aplikasi mobile. Harga berupa angka rupiah.
// Urutan dari yang terlama ke terbaru (id naik); aplikasi mengambil "menu terbaru" dari ujung daftar.
func (h *PublicHandler) Menu(w http.ResponseWriter, r *http.Request) {
	items, err := h.menu.ListAdmin(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	ratings, err := h.reviews.ProductRatings(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	out := make([]menuResponse, 0, len(items))
	for i := len(items) - 1; i >= 0; i-- { // ListAdmin terbaru dulu -> balik
		it := items[i]
		pr := ratings[it.ID]
		out = append(out, menuResponse{
			ID:          it.ID,
			Name:        it.Name,
			Category:    it.Category,
			CategoryID:  categorySlug(it.Category),
			Price:       it.Price,
			PriceLarge:  it.PriceLarge,
			Status:      it.Status,
			Image:       it.Image,
			Description: it.Description,
			Rating:      round1(pr.Average),
			RatingCount: pr.Count,
		})
	}
	httpx.JSON(w, http.StatusOK, out)
}
