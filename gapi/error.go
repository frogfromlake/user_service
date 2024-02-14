package gapi

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

func invalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	badRequest := &errdetails.BadRequest{FieldViolations: violations}
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")

	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}

	return statusDetails.Err()
}

type MultiError []error

func (me MultiError) Error() string {
	var sb strings.Builder
	for _, err := range me {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

func handleDatabaseError(err error) error {
	var pgErr *pgconn.PgError
	var errs MultiError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			errs = append(errs, status.Errorf(codes.AlreadyExists, "unique violation occurred: %v", err))
		case "23503": // foreign_key_violation
			errs = append(errs, status.Errorf(codes.FailedPrecondition, "foreign key violation occurred: %v", err))
		case "23502": // not_null_violation
			errs = append(errs, status.Errorf(codes.InvalidArgument, "not null violation occurred: %v", err))
		case "23514": // check_violation
			errs = append(errs, status.Errorf(codes.OutOfRange, "check violation occurred: %v", err))
		case "2200L": // invalid_text_representation
			errs = append(errs, status.Errorf(codes.InvalidArgument, "invalid text representation: %v", err))
		case "22P02": // invalid_text_representation
			errs = append(errs, status.Errorf(codes.InvalidArgument, "invalid text representation: %v", err))
		case "23P01": // exclusion_violation
			errs = append(errs, status.Errorf(codes.AlreadyExists, "exclusion violation occurred: %v", err))
		case "25006": // read_only_sql_transaction
			errs = append(errs, status.Errorf(codes.PermissionDenied, "read-only SQL transaction: %v", err))
		case "22023": // no_data
			errs = append(errs, status.Errorf(codes.NotFound, "no data: %v", err))
		case "54000": // too_many_connections
			errs = append(errs, status.Errorf(codes.ResourceExhausted, "too many connections: %v", err))
		default:
			errs = append(errs, status.Errorf(codes.Unknown, "unknown database error: %v", err))
		}
	} else {
		errs = append(errs, status.Errorf(codes.Internal, "internal error: %v", err))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
