package gapi

import (
	"errors"
	"fmt"

	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// API Error Handling
// CustomError represents a custom error with a status code and field violation details.
type CustomError struct {
	StatusCode codes.Code
	Violation  *errdetails.BadRequest_FieldViolation
}

// Error returns the string representation of the error.
func (c *CustomError) Error() string {
	return fmt.Sprintf("status %s: %s", c.StatusCode, c.Violation.Description)
}

// WithDetails adds field violation details to a status error.
func (c *CustomError) WithDetails(field string, err error) *CustomError {
	c.Violation = fieldViolation(field, err)
	return c
}

// fieldViolation creates a new field violation with the given field and error.
func fieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

// invalidArgumentError creates a new invalid argument error with the given violations.
func invalidArgumentErrors(violations []*CustomError) error {
	badRequest := &errdetails.BadRequest{FieldViolations: make([]*errdetails.BadRequest_FieldViolation, len(violations))}
	for i, violation := range violations {
		badRequest.FieldViolations[i] = violation.Violation
	}
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")
	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}
	return statusDetails.Err()
}

// invalidArgumentError creates a new invalid argument error with the given violation.
func invalidArgumentError(violation *CustomError) error {
	badRequest := &errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{violation.Violation},
	}
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")
	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}
	return statusDetails.Err()
}

// handleDatabaseError is a function that takes an error and returns a new error with additional details.
func handleDatabaseError(err error) error {
	var pgErr *pgconn.PgError
	var dbErr *pb.DatabaseError // Use the generated DatabaseError struct

	if errors.As(err, &pgErr) {
		dbErr = &pb.DatabaseError{ // Use the generated DatabaseError struct
			Code:        pgErr.Code,
			Message:     pgErr.Message,
			Description: "Database operation failed",
		}

		var statusCode codes.Code
		switch pgErr.Code {
		case "23505": // unique_violation
			statusCode = codes.AlreadyExists
		case "23503": // foreign_key_violation
			statusCode = codes.FailedPrecondition
		case "23502": // not_null_violation
			statusCode = codes.InvalidArgument
		case "23514": // check_violation
			statusCode = codes.OutOfRange
		case "2200L": // invalid_text_representation
			statusCode = codes.InvalidArgument
		case "22P02": // invalid_text_representation
			statusCode = codes.InvalidArgument
		case "23P01": // exclusion_violation
			statusCode = codes.AlreadyExists
		case "25006": // read_only_sql_transaction
			statusCode = codes.PermissionDenied
		case "22023": // no_data
			statusCode = codes.NotFound
		case "54000": // too_many_connections
			statusCode = codes.ResourceExhausted
		default:
			statusCode = codes.Unknown
		}

		// Create a status with the error code and message
		statusDetails := status.New(statusCode, dbErr.Message)

		// Attach the DatabaseError details to the status
		statusDetails, err = statusDetails.WithDetails(dbErr)
		if err != nil {
			// If there's an error attaching the details, return an internal error
			return status.Error(codes.Internal, "internal error: "+err.Error())
		}
		return statusDetails.Err()
	} else {
		// If the error is not a PostgreSQL error, return an internal error
		return status.Error(codes.Internal, "internal error: "+err.Error())
	}
}
