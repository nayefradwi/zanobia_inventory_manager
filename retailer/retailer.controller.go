package retailer

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
)

type RetailerController struct {
	service IRetailerService
}

func NewRetailController(service IRetailerService) RetailerController {
	return RetailerController{
		service,
	}
}

func (c RetailerController) CreateRetailer(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Retailer](w, r.Body, func(retailer Retailer) {
		err := c.service.CreateRetailer(r.Context(), retailer)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Retailer created successfully",
		})
	})
}

func (c RetailerController) AddRetailerContacts(w http.ResponseWriter, r *http.Request) {
	idVal := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idVal)
	common.ParseBody[[]RetailerContact](w, r.Body, func(contacts []RetailerContact) {
		err := c.service.AddRetailerContacts(r.Context(), id, contacts)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Retailer contacts added successfully",
		})
	})
}

func (c RetailerController) AddRetailerContactInfo(w http.ResponseWriter, r *http.Request) {
	idVal := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idVal)
	common.ParseBody[RetailerContact](w, r.Body, func(contact RetailerContact) {
		err := c.service.AddRetailerContactInfo(r.Context(), id, contact)
		common.WriteCreatedResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Retailer contact info added successfully",
		})
	})
}

func (c RetailerController) GetRetailers(w http.ResponseWriter, r *http.Request) {
	retailers, err := c.service.GetRetailers(r.Context())
	common.WriteResponse[common.PaginatedResponse[Retailer]](common.Result[common.PaginatedResponse[Retailer]]{
		Error:  err,
		Writer: w,
		Data:   retailers,
	})
}

func (c RetailerController) GetRetailer(w http.ResponseWriter, r *http.Request) {
	idVal := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idVal)
	retailer, err := c.service.GetRetailer(r.Context(), id)
	common.WriteResponse[Retailer](common.Result[Retailer]{
		Error:  err,
		Writer: w,
		Data:   retailer,
	})
}

func (c RetailerController) RemoveRetailerContactInfo(w http.ResponseWriter, r *http.Request) {
	idVal := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idVal)
	err := c.service.RemoveRetailerContactInfo(r.Context(), id)
	common.WriteEmptyResponse(common.EmptyResult{
		Error:   err,
		Writer:  w,
		Message: "Retailer contact info removed successfully",
	})
}

func (c RetailerController) RemoveRetailer(w http.ResponseWriter, r *http.Request) {
	idVal := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idVal)
	err := c.service.RemoveRetailer(r.Context(), id)
	common.WriteEmptyResponse(common.EmptyResult{
		Error:   err,
		Writer:  w,
		Message: "Retailer removed successfully",
	})
}

func (c RetailerController) UpdateRetailer(w http.ResponseWriter, r *http.Request) {
	common.ParseBody[Retailer](w, r.Body, func(retailer Retailer) {
		err := c.service.UpdateRetailer(r.Context(), retailer)
		common.WriteEmptyResponse(common.EmptyResult{
			Error:   err,
			Writer:  w,
			Message: "Retailer updated successfully",
		})
	})
}
