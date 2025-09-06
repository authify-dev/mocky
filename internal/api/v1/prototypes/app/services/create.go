package services

import (
	"common/domain/customctx"
	"common/domain/logger"
	"common/utils"
	"mocky/internal/api/v1/prototypes/domain/commands"
	prototypes "mocky/internal/db/mongo/prototypes"
	"net/http"
)

func (s *PrototypesService) Create(cc *customctx.CustomContext, prototype commands.CreatePrototypeCommand) utils.Response[prototypes.PrototypeModel] {

	entry := logger.FromContext(cc.Context())

	entry.Info("Creating prototype")

	prototypeEntity := prototype.ToEntity()

	prototypeModel := prototypes.PrototypeModel{
		Request:  prototypeEntity.Request,
		Response: prototypeEntity.Response,
		Name:     prototypeEntity.Name,
	}

	result := s.prototypesRepository.SaveOrUpdate(cc, prototypeModel)
	if result.Err != nil {
		entry.Error(result.Err.Error())
		return utils.Response[prototypes.PrototypeModel]{
			Error:      result.Err,
			StatusCode: http.StatusInternalServerError,
			Success:    false,
		}
	}

	prototypeModel.ID = result.Data

	return utils.Response[prototypes.PrototypeModel]{
		Data:       prototypeModel,
		StatusCode: http.StatusCreated,
		Success:    true,
	}
}
