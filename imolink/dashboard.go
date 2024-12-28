package imolink

import (
	"net/http"

	"encore.app/internal/pkg/apierror"
	"encore.dev/beta/errs"
)

//encore:api public raw path=/dashboard
func (s *Service) Dashboard(w http.ResponseWriter, req *http.Request) {
	if err := s.tmpls.Execute(w, nil); err != nil {
		http.Error(w, "Falha ao renderizar o painel", http.StatusInternalServerError)
	}
}

//encore:api public raw method=POST path=/dashboard/update-assistant
func (s *Service) UpdateAssistantHandler(w http.ResponseWriter, req *http.Request) {
	if err := s.UpdateAssistant(req.Context()); err != nil {
		apierror.E("falha ao atualizar o assistente", err, errs.Internal)
		http.Error(w, "Falha ao atualizar o assistente", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
}
