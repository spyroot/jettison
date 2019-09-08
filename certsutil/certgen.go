package certsutil

type NodeCertRequest struct {
	CN  string `json:"CN"`
	Key struct {
		Algo string `json:"algo"`
		Size int    `json:"size"`
	} `json:"key"`
	Names []struct {
		C  string `json:"C"`
		L  string `json:"L"`
		O  string `json:"O"`
		OU string `json:"OU"`
		ST string `json:"ST"`
	} `json:"names"`
}
