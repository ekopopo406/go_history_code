package controllers

import (
	"go_chat_v1_test/internal/i18n"
	"go_chat_v1_test/internal/middlewares"
	"go_chat_v1_test/internal/services"
	"go_chat_v1_test/internal/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonController struct {
	commonService services.CommonServices
}

func NewCommonController(commonService services.CommonServices) *CommonController {
	return &CommonController{commonService: commonService}
}

func (a *CommonController) SendPhoneCode(v *validator.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.URL.Query().Get("lang")

		var respondMsg string
		var input services.SendPhoneCodeInput
		ctx := c.Request.Context()
		// 解析并验证
		if err := v.ParseAndValidate(c.Request, &input, lang); err != nil {
			// 格式化验证错误
			errors := v.FormatValidationErrors(err, lang)
			middlewares.RespondWithJSON(c, http.StatusBadRequest, errors, "")
			return
		}

		var data = a.commonService.SendPhoneCode(ctx, input)
		respondMsg = i18n.T(lang, "success")
		middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, data)
	}

}
