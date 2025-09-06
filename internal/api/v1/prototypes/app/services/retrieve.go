package services

import (
	"common/domain/customctx"
	"common/domain/logger"
	"common/utils"
	prototypes "mocky/internal/db/mongo/prototypes"
	"net/http"
)

func (s *PrototypesService) Retrieve(cc *customctx.CustomContext, id string) utils.Response[prototypes.PrototypeModel] {

	entry := logger.FromContext(cc.Context())

	entry.Info("Retrieving prototype")

	prototype := s.prototypesRepository.Find(cc.Context(), id)
	if prototype.Err != nil {
		return utils.Response[prototypes.PrototypeModel]{
			Error:      prototype.Err,
			StatusCode: http.StatusInternalServerError,
			Success:    false,
		}
	}

	return utils.Response[prototypes.PrototypeModel]{
		StatusCode: http.StatusOK,
		Data:       prototype.Data,
		Success:    true,
	}
}
