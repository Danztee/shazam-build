package songs

import (
	"log"
	"net/http"

	"github.com/Danztee/shazam-build/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) AddSong(w http.ResponseWriter, r *http.Request) {

	var songPayload createSongPayload
	if err := json.ReadJSON(r, &songPayload); err != nil {
		log.Println("error reading song payload", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.service.AddSong(r.Context(), songPayload)
	if err != nil {
		log.Println("error adding song", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.WriteJSON(w, http.StatusOK, "song added successfully")

}
