package main

import "strings"

type Supervisor interface {
	Supervise()
}

type ServiceSupervisor struct {
	serviceName     string
	errChan         chan error
	fatalErrs       []error
	fatalErrHandler func(err error, serviceName string)
}

func (s ServiceSupervisor) Supervise() {
	for err := range s.errChan {
		if err == nil {
			continue
		}
		for _, fatalErr := range s.fatalErrs {
			if strings.Contains(err.Error(), fatalErr.Error()) {
				s.fatalErrHandler(err, s.serviceName)
			}
		}
	}
}

func newServiceSupervisor(serviceName string, errChan chan error, fatalErrs []error, fatalErrorHandler func(err error, serviceName string)) Supervisor {
	return &ServiceSupervisor{
		serviceName:     serviceName,
		errChan:         errChan,
		fatalErrs:       fatalErrs,
		fatalErrHandler: fatalErrorHandler,
	}
}
