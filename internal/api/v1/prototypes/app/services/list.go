package services

import (
	"common/domain/customctx"
	"common/domain/logger"
	"common/utils"
	"mocky/internal/db/mongo/prototypes"
	"net/http"
)

func (s *PrototypesService) List(cc *customctx.CustomContext) utils.Response[prototypes.PrototypeListModel] {

	entry := logger.FromContext(cc.Context())

	entry.Info("Listing prototypes")

	prototypesList := s.prototypesRepository.FindAll(cc.Context())
	if prototypesList.Err != nil {
		return utils.Response[prototypes.PrototypeListModel]{
			StatusCode: http.StatusInternalServerError,
			Error:      prototypesList.Err,
			Success:    false,
		}
	}

	return utils.Response[prototypes.PrototypeListModel]{
		StatusCode: http.StatusOK,
		Results:    prototypesList.Data,
		Success:    true,
	}
}
