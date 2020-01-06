package featureProc

import (
	"io"

	"github.com/paulmach/orb/geojson"
)

type Processor struct {
	lastSupplier Supplier
	out          io.Writer
	closed       bool
}

type readerSupplier struct {
	reader *ChoppedFeatureReader
}

func (s *readerSupplier) Get() (*geojson.Feature, error) {
	f, _, err := s.reader.Read()
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return f, nil
}

func NewWithReader(in io.Reader, out io.Writer) *Processor {
	return &Processor{
		lastSupplier: &readerSupplier{NewChoppedFeatureReader(in)},
		out:          out,
	}
}

func NewWithSupplier(supplier Supplier, out io.Writer) *Processor {
	return &Processor{lastSupplier: supplier, out: out}
}

func (p *Processor) Map(op func(*geojson.Feature) (*geojson.Feature, error)) *Processor {
	if p.closed {
		panic("Feature processor is closed")
	}
	p.lastSupplier = &mapStep{op: op, source: p.lastSupplier}
	return p
}

func (p *Processor) Filter(op func(*geojson.Feature) (bool, error)) *Processor {
	if p.closed {
		panic("Feature processor is closed")
	}
	p.lastSupplier = &filterStep{op: op, source: p.lastSupplier}
	return p
}

type arraySupplier struct {
	array []*geojson.Feature
}

func (s *arraySupplier) Get() (*geojson.Feature, error) {
	if len(s.array) == 0 {
		return nil, nil
	}
	f := s.array[0]
	s.array = s.array[1:]
	return f, nil
}

func (p *Processor) FlatMap(op func(*geojson.Feature) ([]*geojson.Feature, error)) *Processor {
	if p.closed {
		panic("Feature processor is closed")
	}
	newOp := func(f *geojson.Feature) (Supplier, error) {
		array, err := op(f)
		if err != nil {
			return nil, err
		}
		return &arraySupplier{array}, nil
	}
	p.lastSupplier = &flatMapStep{op: newOp, source: p.lastSupplier}
	return p
}

func (p *Processor) Run() error {
	if p.closed {
		panic("Feature processor is closed")
	}

	writer := NewChoppedFeatureWriter(p.out)
	for {
		f, err := p.lastSupplier.Get()
		if err != nil {
			return err
		}
		if f == nil {
			/* All done */
			p.closed = true
			return nil
		}
		if err = writer.Encode(f); err != nil {
			return err
		}
	}
}
