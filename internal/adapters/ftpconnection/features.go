package ftpconnection

const (
	FeatureMLST = "MLST"
	FeatureMDTM = "MDTM"
	FeatureMFMT = "MFMT"
	FeaturePRET = "PRET"
	FeatureUTF8 = "UTF8"
	FeatureEPSV = "EPSV"
)

type serverFeatures struct {
	supportMLST bool
	supportMDTM bool
	supportMFMT bool
	supportPRET bool
	supportEPSV bool
	supportUTF8 bool
}

func newServerFeatures(featureMap map[string]string) *serverFeatures {
	sf := &serverFeatures{}

	// if _, ok := featureMap[FeatureMLST]; ok && !c.dialOptions.disableMLST {
	//	c.features.supportMLST = true
	// }
	// c.mdtmCanWrite = c.mdtmSupported && c.dialOptions.writingMDTM
	_, sf.supportMLST = featureMap[FeatureMLST]
	_, sf.supportPRET = featureMap[FeaturePRET]
	_, sf.supportMDTM = featureMap[FeatureMDTM]
	_, sf.supportMFMT = featureMap[FeatureMFMT]
	_, sf.supportEPSV = featureMap[FeatureEPSV]
	_, sf.supportUTF8 = featureMap[FeatureUTF8]
	return sf
}
