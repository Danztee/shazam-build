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
		json.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.service.AddSong(r.Context(), songPayload)
	if err != nil {
		log.Println("error adding song", err)
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, map[string]string{"message": "song added successfully"})

}

func (h *handler) MatchSong(w http.ResponseWriter, r *http.Request) {
	var payload matchSongPayload
	if err := json.ReadJSON(r, &payload); err != nil {
		log.Println("error reading match payload", err)
		json.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	song, err := h.service.MatchSong(r.Context(), payload.Data)
	if err != nil {
		log.Println("error matching song", err)
		json.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, song)
}
