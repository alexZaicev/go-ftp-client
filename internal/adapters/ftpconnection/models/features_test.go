package models_test

import (
	"github.com/alexZaicev/go-ftp-client/internal/adapters/ftpconnection/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_NewServerFeatures_Success(t *testing.T) {
	// arrange
	featureMap := map[string]string{
		"MLST": "",
		"MDTM": "",
		"MFMT": "",
		"PRET": "",
		"UTF8": "",
		"EPSV": "",
	}

	// act
	sf := models.NewServerFeatures(featureMap)

	// assert
	require.NotNil(t, sf)
	
	assert.True(t, sf.SupportMLST)
	assert.True(t, sf.SupportMDTM)
	assert.True(t, sf.SupportMFMT)
	assert.True(t, sf.SupportPRET)
	assert.True(t, sf.SupportUTF8)
	assert.True(t, sf.SupportEPSV)
}
