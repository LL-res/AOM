package errs

import "errors"

var (
	NO_SUFFICENT_DATA  = errors.New("no sufficient data")
	UNREADY_TO_PREDICT = errors.New("the model is not ready to predict")
)
