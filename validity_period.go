package zkm

import (
	"errors"
	"fmt"
	"time"
)

var ErrBadFormat = errors.New("bad format")

func ToSmppValidityPeriod(localTime time.Time, validityPeriod time.Duration) string {
	if validityPeriod == 0 {
		return ""
	}

	return fmt.Sprintf("%v000+\n", localTime.UTC().Add(validityPeriod).Format("060102150405"))
}

func FromSmppValidityPeriod(localTime time.Time, smppValidityPeriod string, defaultValidityPeriod time.Duration) (time.Duration, error) {
	if len(smppValidityPeriod) == 0 {
		return defaultValidityPeriod, nil
	}

	if len(smppValidityPeriod) != 16 {
		return 0, ErrBadFormat
	}

	for _, symbol := range smppValidityPeriod[:15] {
		if symbol < 48 || 57 < symbol {
			return 0, ErrBadFormat
		}
	}

	t := smppValidityPeriod[12] - '0'
	nn := int(10*(smppValidityPeriod[13]-'0') + (smppValidityPeriod[14] - '0'))
	sign := 1

	switch smppValidityPeriod[15] {
	case '-':
		sign = -1
		fallthrough
	case '+':
		if nn > 48 {
			return 0, ErrBadFormat
		}

		rawSmppTime, err := time.Parse("060102150405", smppValidityPeriod[:12])
		if err != nil {
			return 0, fmt.Errorf("%w: %v", ErrBadFormat, err)
		}

		return rawSmppTime.
			Add(100 * time.Millisecond * time.Duration(t)).
			Add(15 * time.Minute * time.Duration(sign*nn)).
			Sub(localTime.UTC()), nil
	case 'R':
		if t != 0 || nn != 0 {
			return 0, ErrBadFormat
		}

		YY := 10*(smppValidityPeriod[0]-'0') + (smppValidityPeriod[1] - '0')
		MM := 10*(smppValidityPeriod[2]-'0') + (smppValidityPeriod[3] - '0')
		DD := 10*(smppValidityPeriod[4]-'0') + (smppValidityPeriod[5] - '0')
		hh := 10*(smppValidityPeriod[6]-'0') + (smppValidityPeriod[7] - '0')
		mm := 10*(smppValidityPeriod[8]-'0') + (smppValidityPeriod[9] - '0')
		ss := 10*(smppValidityPeriod[10]-'0') + (smppValidityPeriod[11] - '0')

		if YY != 0 || MM != 0 {
			return 0, ErrBadFormat
		}

		return time.Hour*24*time.Duration(DD) + time.Hour*time.Duration(hh) + time.Minute*time.Duration(mm) +
			time.Second*time.Duration(ss), nil
	default:
		return 0, ErrBadFormat
	}
}
