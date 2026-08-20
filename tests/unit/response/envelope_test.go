package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess_DoesNotHTMLEscapeAmpersand(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	signed := "https://storage.googleapis.com/bucket/obj.jpeg?X-Goog-Algorithm=GOOG4-RSA-SHA256&X-Goog-Credential=sa"

	response.Success(c, http.StatusOK, "ok", gin.H{"url": signed})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "&X-Goog-Credential=")
	assert.NotContains(t, w.Body.String(), `\u0026`)

	var env response.SuccessEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Equal(t, "ok", env.Message)
}

func TestOK_AndCreated(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		response.OK(c, "done", gin.H{"id": 1})
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
	})
	t.Run("created", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		response.Created(c, "created", nil)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
	})
}

func TestError_Envelope(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var env response.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.False(t, env.Success)
	assert.Equal(t, "INVALID_REQUEST", env.Error.Code)
	assert.Equal(t, "invalid request", env.Error.Message)
}

func TestFromError_MappedAndUnmapped(t *testing.T) {
	t.Run("app error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		err := apperrors.NewAppError("VEHICLE_NOT_FOUND", "vehicle not found", http.StatusNotFound, nil)
		response.FromError(c, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"VEHICLE_NOT_FOUND"`)
	})
	t.Run("unmapped", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		response.FromError(c, errors.New("boom"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), apperrors.CodeInternal)
	})
}
