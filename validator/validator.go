package validator

import (
	"fmt"
	"indicartcoin/structs"
)

// selectValidator selects a validator based on their stake, rewards, and penalties.
func SelectValidator(validators []structs.Validator) (structs.Validator, error) {
	// Calculate the total score
	MaxScore := float64(0)
	for _, validator := range validators {
		if validator.Stake > MaxScore {
			MaxScore = validator.Stake
		}

	}

	for _, validator := range validators {
		if validator.Stake == MaxScore {
			return validator, nil
		}

	}

	// This should never happen
	return structs.Validator{}, fmt.Errorf("unexpected error in selecting validator")
}
