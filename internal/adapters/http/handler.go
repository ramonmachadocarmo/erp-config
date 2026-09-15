package httpadapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"erp-schema/model"
	"erp/pkg/httpserver"
	"erp/services/config-service/internal/application"
	"erp/services/config-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// personIn shadows domain.Person's promoted Addresses field with the shared erp-schema
// type, so address validation (required fields) runs against the single cross-service
// schema before mapping down into config-service's own Address.
type personIn struct {
	domain.Person
	Addresses []model.Address `json:"addresses"`
}

type Handler struct {
	svc *application.Service
}

func New(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine, jwt gin.HandlerFunc) {
	api := r.Group("/", jwt)
	api.GET("/units", h.list)
	api.POST("/units", h.create)
	api.GET("/units/:id", h.get)
	api.PUT("/units/:id", h.update)
	api.DELETE("/units/:id", h.delete)
	api.GET("/customers", h.listCustomers)
	api.POST("/customers", h.createCustomer)
	api.GET("/customers/:id", h.getCustomer)
	api.PUT("/customers/:id", h.updateCustomer)
	api.DELETE("/customers/:id", h.deleteCustomer)
	api.GET("/suppliers", h.listSuppliers)
	api.POST("/suppliers", h.createSupplier)
	api.GET("/suppliers/:id", h.getSupplier)
	api.PUT("/suppliers/:id", h.updateSupplier)
	api.DELETE("/suppliers/:id", h.deleteSupplier)
	api.GET("/settings", h.listSettings)
	api.GET("/settings/:key", h.getSetting)
	api.PUT("/settings/:key", h.putSetting)
	api.GET("/payment-methods", h.listMethods)
	api.POST("/payment-methods", h.createMethod)
	api.GET("/payment-methods/:id", h.getMethod)
	api.PUT("/payment-methods/:id", h.updateMethod)
	api.GET("/payment-terms", h.listTerms)
	api.POST("/payment-terms", h.createTerm)
	api.GET("/payment-terms/:id", h.getTerm)
	api.PUT("/payment-terms/:id", h.updateTerm)
	api.GET("/cep/:cep", h.lookupCEP)
	api.GET("/geo", h.lookupGeo)
	api.GET("/centers", h.listCenters)
	api.POST("/centers", h.createCenter)
	api.GET("/centers/:id", h.getCenter)
	api.PUT("/centers/:id", h.updateCenter)
	api.GET("/vehicles", h.listVehicles)
	api.POST("/vehicles", h.createVehicle)
	api.GET("/vehicles/:id", h.getVehicle)
	api.PUT("/vehicles/:id", h.updateVehicle)
	api.GET("/company", h.getCompany)
	api.PUT("/company", h.updateCompany)
}

func (h *Handler) list(c *gin.Context) {
	out, err := h.svc.ListUnits(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) create(c *gin.Context) {
	var in domain.UnitOfMeasure
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	out, err := h.svc.CreateUnit(c.Request.Context(), in)
	if errors.Is(err, domain.ErrConflict) {
		httpserver.Error(c, http.StatusConflict, err)
		return
	}
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *Handler) get(c *gin.Context) {
	out, err := h.svc.GetUnit(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) update(c *gin.Context) {
	var in domain.UnitOfMeasure
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.ID = c.Param("id")
	if err := h.svc.UpdateUnit(c.Request.Context(), in); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, domain.ErrConflict) {
			status = http.StatusConflict
		}
		httpserver.Error(c, status, err)
		return
	}
	c.JSON(http.StatusOK, in)
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.svc.DeleteUnit(c.Request.Context(), c.Param("id")); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, domain.ErrInUse) || errors.Is(err, domain.ErrConflict) {
			status = http.StatusConflict
		}
		httpserver.Error(c, status, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listCustomers(c *gin.Context) {
	h.listPeople(c, h.svc.ListCustomers)
}

func (h *Handler) createCustomer(c *gin.Context) {
	h.createPerson(c, h.svc.CreateCustomer)
}

func (h *Handler) getCustomer(c *gin.Context) {
	h.getPerson(c, h.svc.GetCustomer)
}

func (h *Handler) updateCustomer(c *gin.Context) {
	h.updatePerson(c, h.svc.UpdateCustomer)
}

func (h *Handler) deleteCustomer(c *gin.Context) {
	if err := h.svc.DeleteCustomer(c.Request.Context(), c.Param("id")); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, domain.ErrInUse) || errors.Is(err, domain.ErrConflict) {
			status = http.StatusConflict
		}
		httpserver.Error(c, status, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listSuppliers(c *gin.Context) {
	h.listPeople(c, h.svc.ListSuppliers)
}

func (h *Handler) createSupplier(c *gin.Context) {
	h.createPerson(c, h.svc.CreateSupplier)
}

func (h *Handler) getSupplier(c *gin.Context) {
	h.getPerson(c, h.svc.GetSupplier)
}

func (h *Handler) updateSupplier(c *gin.Context) {
	h.updatePerson(c, h.svc.UpdateSupplier)
}

func (h *Handler) deleteSupplier(c *gin.Context) {
	if err := h.svc.DeleteSupplier(c.Request.Context(), c.Param("id")); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, domain.ErrInUse) || errors.Is(err, domain.ErrConflict) {
			status = http.StatusConflict
		}
		httpserver.Error(c, status, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listPeople(c *gin.Context, fn func(context.Context) ([]domain.Person, error)) {
	out, err := fn(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) createPerson(c *gin.Context, fn func(context.Context, domain.Person) (domain.Person, error)) {
	var in personIn
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.Person.Addresses = toDomainAddresses(in.Addresses)
	out, err := fn(c.Request.Context(), in.Person)
	if errors.Is(err, domain.ErrInvalid) {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	if errors.Is(err, domain.ErrConflict) {
		httpserver.Error(c, http.StatusConflict, err)
		return
	}
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *Handler) getPerson(c *gin.Context, fn func(context.Context, string) (domain.Person, error)) {
	out, err := fn(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) updatePerson(c *gin.Context, fn func(context.Context, domain.Person) error) {
	var in personIn
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.Person.Addresses = toDomainAddresses(in.Addresses)
	in.Person.ID = c.Param("id")
	if err := fn(c.Request.Context(), in.Person); err != nil {
		status := http.StatusNotFound
		if errors.Is(err, domain.ErrInvalid) {
			status = http.StatusBadRequest
		} else if errors.Is(err, domain.ErrConflict) {
			status = http.StatusConflict
		}
		httpserver.Error(c, status, err)
		return
	}
	c.JSON(http.StatusOK, in.Person)
}

func (h *Handler) listSettings(c *gin.Context) {
	out, err := h.svc.ListSettings(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) getSetting(c *gin.Context) {
	out, err := h.svc.GetSetting(c.Request.Context(), c.Param("key"))
	if err != nil {
		httpserver.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) putSetting(c *gin.Context) {
	var in domain.Setting
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.Key = c.Param("key")
	if err := h.svc.PutSetting(c.Request.Context(), in); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, domain.ErrNotFound) {
			status = http.StatusNotFound
		}
		httpserver.Error(c, status, err)
		return
	}
	c.JSON(http.StatusOK, in)
}

func (h *Handler) listMethods(c *gin.Context) {
	out, err := h.svc.ListPaymentMethods(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) createMethod(c *gin.Context) {
	var in domain.PaymentMethod
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	out, err := h.svc.CreatePaymentMethod(c.Request.Context(), in)
	h.writePayment(c, out, err, true)
}

func (h *Handler) getMethod(c *gin.Context) {
	out, err := h.svc.GetPaymentMethod(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) updateMethod(c *gin.Context) {
	var in domain.PaymentMethod
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.ID = c.Param("id")
	if err := h.svc.UpdatePaymentMethod(c.Request.Context(), in); err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, in)
}

func (h *Handler) listTerms(c *gin.Context) {
	out, err := h.svc.ListPaymentTerms(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) createTerm(c *gin.Context) {
	var in domain.PaymentTerm
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	out, err := h.svc.CreatePaymentTerm(c.Request.Context(), in)
	h.writePayment(c, out, err, true)
}

func (h *Handler) getTerm(c *gin.Context) {
	out, err := h.svc.GetPaymentTerm(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) updateTerm(c *gin.Context) {
	var in domain.PaymentTerm
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.ID = c.Param("id")
	if err := h.svc.UpdatePaymentTerm(c.Request.Context(), in); err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, in)
}

func (h *Handler) writePayment(c *gin.Context, out any, err error, created bool) {
	if err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	if created {
		c.JSON(http.StatusCreated, out)
		return
	}
	c.JSON(http.StatusOK, out)
}

func paymentStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) lookupCEP(c *gin.Context) {
	out, err := h.svc.LookupCEP(c.Request.Context(), c.Param("cep"))
	if err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) lookupGeo(c *gin.Context) {
	street, city := c.Query("street"), c.Query("city")
	if street != "" || city != "" {
		out, err := h.svc.SearchPlace(c.Request.Context(), street, c.Query("number"), c.Query("district"), city, c.Query("state"), c.Query("zip"))
		if err != nil {
			httpserver.Error(c, paymentStatus(err), err)
			return
		}
		c.JSON(http.StatusOK, out)
		return
	}
	if q := c.Query("q"); q != "" {
		out, err := h.svc.SearchGeo(c.Request.Context(), q)
		if err != nil {
			httpserver.Error(c, paymentStatus(err), err)
			return
		}
		c.JSON(http.StatusOK, out)
		return
	}
	lat, err1 := parseFloat(c.Query("lat"))
	lng, err2 := parseFloat(c.Query("lng"))
	if err1 != nil || err2 != nil {
		httpserver.Error(c, http.StatusBadRequest, domain.ErrInvalid)
		return
	}
	out, err := h.svc.LookupGeo(c.Request.Context(), lat, lng)
	if err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) listCenters(c *gin.Context) {
	out, err := h.svc.ListCenters(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) createCenter(c *gin.Context) {
	var in domain.DistributionCenter
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	out, err := h.svc.CreateCenter(c.Request.Context(), in)
	h.writePayment(c, out, err, true)
}

func (h *Handler) getCenter(c *gin.Context) {
	out, err := h.svc.GetCenter(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) updateCenter(c *gin.Context) {
	var in domain.DistributionCenter
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.ID = c.Param("id")
	if err := h.svc.UpdateCenter(c.Request.Context(), in); err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, in)
}

func (h *Handler) listVehicles(c *gin.Context) {
	out, err := h.svc.ListVehicles(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) createVehicle(c *gin.Context) {
	var in domain.DeliveryVehicle
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	out, err := h.svc.CreateVehicle(c.Request.Context(), in)
	h.writePayment(c, out, err, true)
}

func (h *Handler) getVehicle(c *gin.Context) {
	out, err := h.svc.GetVehicle(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) updateVehicle(c *gin.Context) {
	var in domain.DeliveryVehicle
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	in.ID = c.Param("id")
	if err := h.svc.UpdateVehicle(c.Request.Context(), in); err != nil {
		httpserver.Error(c, paymentStatus(err), err)
		return
	}
	c.JSON(http.StatusOK, in)
}

func (h *Handler) getCompany(c *gin.Context) {
	out, err := h.svc.GetCompany(c.Request.Context())
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) updateCompany(c *gin.Context) {
	var in domain.Company
	if err := c.ShouldBindJSON(&in); err != nil {
		httpserver.Error(c, http.StatusBadRequest, err)
		return
	}
	out, err := h.svc.UpdateCompany(c.Request.Context(), in)
	if err != nil {
		httpserver.Error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func parseFloat(s string) (float64, error) {
	var n float64
	_, err := fmt.Sscan(s, &n)
	return n, err
}
