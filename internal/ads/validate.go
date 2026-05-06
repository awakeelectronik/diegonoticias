package ads

import (
	"fmt"
	"strings"
)

func Validate(b Banner, all []Banner, updatingID string) error {
	if strings.TrimSpace(b.Title) == "" || len(b.Title) > 80 {
		return fmt.Errorf("%w: título requerido (1..80)", ErrInvalid)
	}
	if b.Slot != 1 && b.Slot != 2 {
		return fmt.Errorf("%w: slot inválido", ErrInvalid)
	}
	if strings.TrimSpace(b.ImagePath) == "" {
		return fmt.Errorf("%w: imagePath requerido", ErrInvalid)
	}

	total := 0
	active := 0
	activeInSlot := 0
	for _, x := range all {
		if x.ID == updatingID {
			continue
		}
		total++
		if x.Active {
			active++
			if x.Slot == b.Slot {
				activeInSlot++
			}
		}
	}
	if total >= 7 {
		return fmt.Errorf("%w: máximo 7 banners", ErrInvalid)
	}
	if b.Active && active >= 2 {
		return fmt.Errorf("%w: máximo 2 banners activos", ErrInvalid)
	}
	if b.Active && activeInSlot >= 1 {
		return fmt.Errorf("%w: ya hay un banner activo en ese slot", ErrInvalid)
	}
	return nil
}

