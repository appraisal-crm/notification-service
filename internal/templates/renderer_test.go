package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_RenderAll(t *testing.T) {
	r, err := NewRenderer()
	require.NoError(t, err)

	t.Run("RequestCreatedClient_RU", func(t *testing.T) {
		objType := "apartment"
		addr := "г. Москва, ул. Арбат, д. 10"
		subj, body, err := r.RenderRequestCreatedClient(LocaleRU, "a1b2c3d4-5678-90ef", "test@client.com", "+79991234567", &objType, &addr, "http://localhost:5173")
		require.NoError(t, err)
		assert.Contains(t, subj, "a1b2c3d4")
		assert.Contains(t, subj, "принята")
		assert.Contains(t, body, "Квартира")
		assert.Contains(t, body, "Арбат")
		assert.Contains(t, body, "http://localhost:5173/")
	})

	t.Run("RequestCreatedClient_EN", func(t *testing.T) {
		objType := "apartment"
		addr := "10 Arbat street, Moscow"
		subj, body, err := r.RenderRequestCreatedClient(LocaleEN, "a1b2c3d4-5678-90ef", "test@client.com", "+79991234567", &objType, &addr, "http://localhost:5173")
		require.NoError(t, err)
		assert.Contains(t, subj, "a1b2c3d4")
		assert.Contains(t, subj, "received")
		assert.Contains(t, body, "Apartment")
		assert.Contains(t, body, "Open Client Portal")
		assert.Contains(t, body, "http://localhost:5173/")
	})

	t.Run("RequestCreatedStaff_EN", func(t *testing.T) {
		objType := "land"
		addr := "Moscow region"
		subj, body, err := r.RenderRequestCreatedStaff(LocaleEN, "a1b2c3d4-5678-90ef", "client@mail.ru", "+79991234567", &objType, &addr, "http://localhost:5174")
		require.NoError(t, err)
		assert.Contains(t, subj, "New appraisal request")
		assert.Contains(t, body, "client@mail.ru")
		assert.Contains(t, body, "Land plot")
		assert.Contains(t, body, "http://localhost:5174/requests")
	})

	t.Run("InspectionScheduledClient_EN", func(t *testing.T) {
		subj, body, err := r.RenderInspectionScheduledClient(LocaleEN, "b2c3d4e5-1111", "http://localhost:5173")
		require.NoError(t, err)
		assert.Contains(t, subj, "Property inspection scheduled")
		assert.Contains(t, body, "Check Request Status")
	})

	t.Run("InspectionCompleted_EN", func(t *testing.T) {
		subj, body, err := r.RenderInspectionCompleted(LocaleEN, "b2c3d4e5-1111", 5, "http://localhost:5173", true)
		require.NoError(t, err)
		assert.Contains(t, subj, "completed")
		assert.Contains(t, body, "5 files")
	})

	t.Run("ReportReadyClient_EN", func(t *testing.T) {
		subj, body, err := r.RenderReportReadyClient(LocaleEN, "c3d4e5f6-2222", "http://localhost:5173")
		require.NoError(t, err)
		assert.Contains(t, subj, "report is ready")
		assert.Contains(t, body, "Download Report")
	})
}
