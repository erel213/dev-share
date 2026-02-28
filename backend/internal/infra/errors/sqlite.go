package errors

import (
	"database/sql"
	"net/http"

	pkgerrors "backend/pkg/errors"

	"modernc.org/sqlite"
)

// SQLite extended result codes we care about.
const (
	sqliteConstraintUnique     = 2067
	sqliteConstraintForeignKey = 787
	sqliteConstraintNotNull    = 1299
)

// WrapSQLiteError maps SQLite errors to *pkgerrors.Error.
func WrapSQLiteError(err error, operation string) *pkgerrors.Error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return pkgerrors.WithCode(pkgerrors.CodeNotFound, "record not found").
			WithMetadata("operation", operation).
			WithSeverity(pkgerrors.SeverityWarning)
	}

	if sqliteErr, ok := err.(*sqlite.Error); ok {
		return wrapSQLiteErr(sqliteErr, operation)
	}

	return pkgerrors.Wrap(err, "database operation failed").
		WithMetadata("operation", operation).
		WithHTTPStatus(http.StatusInternalServerError).
		WithSeverity(pkgerrors.SeverityError)
}

func wrapSQLiteErr(err *sqlite.Error, operation string) *pkgerrors.Error {
	base := pkgerrors.Wrap(err, "sqlite error").
		WithMetadata("operation", operation).
		WithMetadata("sqlite_code", err.Code())

	switch int(err.Code()) {
	case sqliteConstraintUnique:
		return base.
			WithCode(pkgerrors.CodeConflict).
			WithHTTPStatus(http.StatusConflict).
			WithSeverity(pkgerrors.SeverityWarning)

	case sqliteConstraintForeignKey:
		return base.
			WithCode(pkgerrors.CodeInvalidInput).
			WithHTTPStatus(http.StatusBadRequest).
			WithSeverity(pkgerrors.SeverityWarning)

	case sqliteConstraintNotNull:
		return base.
			WithCode(pkgerrors.CodeInvalidInput).
			WithHTTPStatus(http.StatusBadRequest).
			WithSeverity(pkgerrors.SeverityWarning)

	default:
		return base.
			WithCode(pkgerrors.CodeDatabase).
			WithHTTPStatus(http.StatusInternalServerError).
			WithSeverity(pkgerrors.SeverityError)
	}
}
