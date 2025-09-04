package chttp

import (
	"bytes"
	"common/domain/customctx"
	"common/domain/logger"
	chttp_errs "common/infrastructure/chttp/errs"
	"common/utils"
	"common/utils/cerrs"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func (c *CustomApiClient) POST(
	ctx *customctx.CustomContext,
	request CustomApiRequest,
) utils.Result[CustomApiResponse] {

	url := c.base_url + request.Path
	entry := logger.FromContext(ctx.Context())
	entry.Info("Make POST request to ", url)

	var jsonBody []byte
	var err error

	if request.Payload != nil {
		// Convert requestBody to JSON
		jsonBody, err = json.Marshal(request.Payload)
		if err != nil {
			entry.Error("Error by turning the body of the application to Json", err)
			return utils.Result[CustomApiResponse]{
				Err: ctx.NewError(chttp_errs.NewMarshalError()),
			}
		}
	}

	if request.Debug {
		logger.InfoD(entry, "JSON Body: ", string(jsonBody))
		logger.InfoD(entry, "Headers: ", request.Headers)
	}

	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		entry.Error("Error creating request", err)
		return utils.Result[CustomApiResponse]{
			Err: ctx.NewError(chttp_errs.NewMakingRequestError(chttp_errs.RequestOptions{
				Method: http.MethodPost,
				URL:    url,
			})),
		}
	}

	c.setHeaders(req, request.Headers)

	resp, err := c.http_client.Do(req)
	if err != nil {
		entry.Error("Error making POST request", err)
		return utils.Result[CustomApiResponse]{
			Err: ctx.NewError(chttp_errs.NewDoingRequestError(chttp_errs.RequestOptions{
				Method: http.MethodPost,
				URL:    url,
			})),
		}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		entry.Error("Error reading the response", err)
		return utils.Result[CustomApiResponse]{
			Err: ctx.NewError(chttp_errs.NewReadingResponseError(chttp_errs.RequestOptions{
				Method: http.MethodPost,
				URL:    url,
			})),
		}
	}

	if request.Debug {
		logger.InfoD(entry, "Response Body: ", string(body))
	}

	var _err *cerrs.CustomError

	if resp.StatusCode != request.ExpectedCode {
		entry.Errorf("Error inesperado: %d — %s", resp.StatusCode, string(body))
		_err = &cerrs.CustomError{
			Code:    resp.StatusCode,
			Message: "error inesperado",
			Scope:   "konectus.knhttp.POST",
		}
	}

	var result map[string]any

	err = json.Unmarshal(body, &result)
	if err != nil {
		entry.Errorf("Error al decodificar JSON: %v", err)
		_err = &cerrs.CustomError{
			Code:    resp.StatusCode,
			Message: "error al decodificar JSON",
			Scope:   "konectus.knhttp.POST",
		}
	}

	if _err != nil {
		return utils.Result[CustomApiResponse]{
			Data: CustomApiResponse{
				Data:       result,
				StatusCode: strconv.Itoa(resp.StatusCode),
			},
			Err: ctx.NewError(_err),
		}
	}

	return utils.Result[CustomApiResponse]{
		Data: CustomApiResponse{
			Data:       result,
			StatusCode: strconv.Itoa(resp.StatusCode),
		},
	}
}
