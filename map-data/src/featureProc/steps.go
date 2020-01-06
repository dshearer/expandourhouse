package featureProc

import "github.com/paulmach/orb/geojson"

type Supplier interface {
	Get() (*geojson.Feature, error)
}

type mapStep struct {
	op     func(d *geojson.Feature) (*geojson.Feature, error)
	source Supplier
}

func (s *mapStep) Get() (*geojson.Feature, error) {
	// get next input
	d, err := s.source.Get()
	if err != nil {
		/* Error */
		return nil, err
	} else if d == nil {
		/* End of data */
		return nil, nil
	}

	// do op
	return s.op(d)
}

type flatMapStep struct {
	op         func(d *geojson.Feature) (Supplier, error)
	source     Supplier
	currOutput Supplier
}

func (s *flatMapStep) Get() (*geojson.Feature, error) {
do:
	for s.currOutput == nil {
		// get next input
		d, err := s.source.Get()
		if err != nil {
			/* Error */
			return nil, err
		} else if d == nil {
			/* End of data */
			return nil, nil
		}

		// do op
		s.currOutput, err = s.op(d)
		if err != nil {
			/* Error */
			return nil, err
		}
	}

	result, err := s.currOutput.Get()
	if err != nil {
		return nil, err
	} else if result == nil {
		/* currOutput is exhausted */
		s.currOutput = nil
		goto do
	}
	return result, nil
}

type filterStep struct {
	op     func(d *geojson.Feature) (bool, error)
	source Supplier
}

func (s *filterStep) Get() (*geojson.Feature, error) {
	for {
		// get next input
		d, err := s.source.Get()
		if err != nil {
			/* Error */
			return nil, err
		} else if d == nil {
			/* End of data */
			return nil, nil
		}

		// do op
		keep, err := s.op(d)
		if err != nil {
			return nil, err
		}

		if keep {
			return d, nil
		}
	}
}
