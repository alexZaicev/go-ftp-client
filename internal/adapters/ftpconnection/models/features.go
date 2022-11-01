package models

const (
	FeatureMLST = "MLST"
	FeatureMDTM = "MDTM"
	FeatureMFMT = "MFMT"
	FeaturePRET = "PRET"
	FeatureUTF8 = "UTF8"
	FeatureEPSV = "EPSV"
	FeatureAUTH = "AUTH"
)

type ServerFeatures struct {
	SupportMLST bool
	SupportMDTM bool
	SupportMFMT bool
	SupportPRET bool
	SupportEPSV bool
	SupportUTF8 bool
	AuthTLS     bool
}

func NewServerFeatures(featureMap map[string]string) *ServerFeatures {
	sf := &ServerFeatures{}
	// FIXME: add support for MLST
	// if _, ok := featureMap[FeatureMLST]; ok && !c.dialOptions.disableMLST {
	//	c.features.supportMLST = true
	// }
	// c.mdtmCanWrite = c.mdtmSupported && c.dialOptions.writingMDTM
	_, sf.SupportMLST = featureMap[FeatureMLST]
	_, sf.SupportPRET = featureMap[FeaturePRET]
	_, sf.SupportMDTM = featureMap[FeatureMDTM]
	_, sf.SupportMFMT = featureMap[FeatureMFMT]
	_, sf.SupportEPSV = featureMap[FeatureEPSV]
	_, sf.SupportUTF8 = featureMap[FeatureUTF8]

	if mode, ok := featureMap[FeatureAUTH]; ok && mode == "TLS" {
		sf.AuthTLS = true
	}

	return sf
}
