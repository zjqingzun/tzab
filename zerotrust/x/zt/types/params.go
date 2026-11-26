package types

const DefaultAdmin string = "admin"
const DefaultGateway string = "gateway"
const DefaultMaxLogBatch uint64 = 0

// NewParams creates a new Params instance.
func NewParams(admin string, gateway string, max_log_batch uint64) Params {
	return Params{Admin: admin,
		Gateway: gateway,

		MaxLogBatch: max_log_batch,

		// DefaultParams returns a default set of parameters.
	}
}

func DefaultParams() Params {
	return NewParams(DefaultAdmin, DefaultGateway, DefaultMaxLogBatch)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if err := validateAdmin(p.Admin); err != nil {
		return err
	}
	if err := validateGateway(p.Gateway); err != nil {
		return err
	}
	if err := validateMaxLogBatch(p.MaxLogBatch); err != nil {
		return err
	}

	return nil
}
func validateAdmin(v string) error {

	return nil
}
func validateGateway(v string) error {

	return nil
}
func validateMaxLogBatch(v uint64) error {

	return nil
}
