package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/zhashkevych/creatly-backend/internal/domain"
	"github.com/zhashkevych/creatly-backend/internal/service"
	"net/http"
)

// @Summary Admin Get Survey
// @Security AdminAuth
// @Tags admins-surveys
// @Description admin get survey
// @ModuleID adminGetSurvey
// @Accept  json
// @Produce  json
// @Param id path string true "module id"
// @Success 200 {object} domain.Survey
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /admins/modules/{id}/survey [get]
func (h *Handler) adminGetSurvey(c *gin.Context) {
	id, err := parseIdFromPath(c, "id")
	if err != nil {
		newResponse(c, http.StatusBadRequest, "invalid id param")

		return
	}

	module, err := h.services.Modules.GetById(c.Request.Context(), id)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, "failed to get module")

		return
	}

	c.JSON(http.StatusOK, module.Survey)
}

type createSurveyInput struct {
	Title     string     `json:"title" binding:"required"`
	Required  bool       `json:"required"`
	Questions []question `json:"questions"`
}

type question struct {
	Question      string   `json:"question" binding:"required"`
	AnswerType    string   `json:"answerType" binding:"required"`
	AnswerOptions []string `json:"answerOptions"`
}

// @Summary Admin Create/Update Survey
// @Security AdminAuth
// @Tags admins-surveys
// @Description admin create/update survey
// @ModuleID adminCreateSurvey
// @Accept  json
// @Produce  json
// @Param id path string true "module id"
// @Param input body createSurveyInput true "survey info"
// @Success 201 {string} ok
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /admins/modules/{id}/survey [post]
func (h *Handler) adminCreateOrUpdateSurvey(c *gin.Context) {
	id, err := parseIdFromPath(c, "id")
	if err != nil {
		newResponse(c, http.StatusBadRequest, "invalid id param")

		return
	}

	var inp createSurveyInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	if err := h.services.Surveys.Create(c.Request.Context(), service.CreateSurveyInput{
		ModuleID: id,
		SchoolID: school.ID,
		Survey: domain.Survey{
			Title:     inp.Title,
			Required:  inp.Required,
			Questions: toQuestions(inp.Questions),
		},
	}); err != nil {
		newResponse(c, http.StatusInternalServerError, "invalid input body")

		return
	}

	c.Status(http.StatusCreated)
}

// @Summary Admin Delete Survey
// @Security AdminAuth
// @Tags admins-surveys
// @Description admin delete survey
// @ModuleID adminDeleteSurvey
// @Accept  json
// @Produce  json
// @Param id path string true "module id"
// @Success 200 {object} domain.Survey
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /admins/modules/{id}/survey [delete]
func (h *Handler) adminDeleteSurvey(c *gin.Context) {
	id, err := parseIdFromPath(c, "id")
	if err != nil {
		newResponse(c, http.StatusBadRequest, "invalid id param")

		return
	}

	school, err := getSchoolFromContext(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	if err := h.services.Surveys.Delete(c.Request.Context(), school.ID, id); err != nil {
		newResponse(c, http.StatusInternalServerError, "failed to delete survey")

		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) adminGetSurveyResults(c *gin.Context) {

}

func (h *Handler) adminGetSurveyStudentResults(c *gin.Context) {

}

func toQuestions(qs []question) []domain.SurveyQuestion {
	res := make([]domain.SurveyQuestion, len(qs))

	for i := range qs {
		res[i] = domain.SurveyQuestion{
			Question:      qs[i].Question,
			AnswerType:    qs[i].AnswerType,
			AnswerOptions: qs[i].AnswerOptions,
		}
	}

	return res
}
