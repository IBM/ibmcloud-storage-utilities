package logger

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/crn"
	uid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	//CrnLabel is label in log for the crn
	CrnLabel = "crn"

	// PodNameEnvVar is the pod name environment variable
	PodNameEnvVar = "POD_NAME"

	//PodName is the zap field key label for pod name
	PodName = "podName"

	// RequestIDLabel is the context key for storing the request ID
	RequestIDLabel = "requestID"

	// TriggerKeyLabel is the context key for storing the trigger key
	TriggerKeyLabel = "triggerKey"
)

// ZapLogger is the global logger
var ZapLogger *zap.Logger

// GetZapLogger returns an instance of the logger, initializing a new logger
func GetZapLogger() (*zap.Logger, error) {
	if ZapLogger == nil {
		return NewZapLogger()
	}
	return ZapLogger, nil
}

// GetZapContextLogger Creates a new logger based from the global logger and adds values from the
// context as logging fields. If the context passed in is null then it
// returns the global logger
func GetZapContextLogger(ctx context.Context) (*zap.Logger, error) {
	var contextLogger *zap.Logger
	globalLogger, _ := GetZapLogger()
	if ctx != nil {
		contextLogger = addContextFields(ctx, globalLogger)
		return contextLogger, nil
	}
	return globalLogger, nil
}

// GetZapContextLogger Creates a new logger based from the global logger and adds RequestID from the
// context as logging field.
func GetZapDefaultContextLogger() (*zap.Logger, error) {
	var contextLogger *zap.Logger
	globalLogger, _ := GetZapLogger()
	contextLogger = addContextFields(generateContextWithRequestID(), globalLogger)
	return contextLogger, nil
}

//GetZapContextLoggerFromLogger creates a new logger based from an existing logger and adds values from the
//context as logging fields. If the context passed in is null then it
//returns the global logger
func GetZapContextLoggerFromLogger(ctx context.Context, origLogger *zap.Logger) (*zap.Logger, error) {
	var contextLogger *zap.Logger
	if origLogger == nil {
		return origLogger, errors.New("a valid logger needs to be passed in")
	}
	if ctx != nil {
		contextLogger = addContextFields(ctx, origLogger)
		return contextLogger, nil
	}
	return origLogger, nil
}

// Adds fields to an existing logger using values in the context
func addContextFields(ctx context.Context, origLogger *zap.Logger) *zap.Logger {
	if _, ok := ctx.Value(TriggerKeyLabel).(string); ok {
		origLogger = origLogger.With(CreateZapTiggerKeyField(ctx))
	}
	if _, ok := ctx.Value(RequestIDLabel).(string); ok {
		origLogger = origLogger.With(CreateZapRequestIDField(ctx))
	}
	return origLogger
}

// NewZapLogger creates and returns a new global logger. It overwrites the
// existing global logger if that has been previously defined.
func NewZapLogger() (*zap.Logger, error) {
	productionConfig := zap.NewProductionConfig()
	productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	lgr, _ := productionConfig.Build()
	lgr, _ = CreateZapCRNLogger(lgr)
	ZapLogger = lgr
	return ZapLogger, nil
}

// CreateZapCRNLogger takes a zap logger and adds a crn field
// NOTE: the logger returned is a different logger from the one passed in
func CreateZapCRNLogger(logger *zap.Logger) (*zap.Logger, error) {
	serviceCRN, err := crn.GetServiceCRN()
	if err != nil {
		logger.Error("Error when retrieving the CRN information.", zap.Error(err))
		return logger, nil
	}
	return logger.With(zapcore.Field{Key: CrnLabel, Type: zapcore.StringType, String: fmt.Sprintf("%s:log", serviceCRN)}), nil
}

// CreatePodNameLogger takes a zap logger and adds a pod name field
// NOTE: the logger returned is a different logger from the one passed in
func CreatePodNameLogger(logger *zap.Logger) (*zap.Logger, error) {
	if logger == nil {
		return nil, errors.New("logger passed in can not be null")
	}
	podNameField := CreateZapPodNameKeyField()
	return logger.With(podNameField), nil
}

//CreateZapRequestIDField Creates a zap logger field containing the request ID, convenience method for creating the
//field in cases where the ContextLogger can't be used and the field needs to be passed
//in as a parameter in the logging statements
func CreateZapRequestIDField(ctx context.Context) zapcore.Field {
	if ctx != nil {
		if requestID, ok := ctx.Value(RequestIDLabel).(string); ok {
			return zapcore.Field{Key: RequestIDLabel, Type: zapcore.StringType, String: requestID}
		}
	}
	return zapcore.Field{Key: RequestIDLabel, Type: zapcore.StringType, String: ""}
}

//CreateZapTiggerKeyField Creates a zap logger field containing the trigger key for a job, convenience method for creating the
//field in cases where the ContextLogger can't be used and the field needs to be passed
//in as a parameter in the logging statements
func CreateZapTiggerKeyField(ctx context.Context) zapcore.Field {
	if ctx != nil {
		if triggerKey, ok := ctx.Value(TriggerKeyLabel).(string); ok {
			return zapcore.Field{Key: TriggerKeyLabel, Type: zapcore.StringType, String: triggerKey}
		}
	}
	return zapcore.Field{Key: TriggerKeyLabel, Type: zapcore.StringType, String: ""}
}

//CreateZapPodNameKeyField Creates a zap logger field containing the pod name that the container is in,
// convenience method for creating the field so it can be passed
//in as a parameter in the logging statements
func CreateZapPodNameKeyField() zapcore.Field {
	pod := os.Getenv(PodNameEnvVar)
	// if the pod name isn't set then the value will be empty
	return zapcore.Field{Key: PodName, Type: zapcore.StringType, String: pod}
}

// Creates a context that contains a unique request ID
func generateContextWithRequestID() context.Context {
	//	requestID := uid.NewV4().String()
	req_uuid, _ := uid.NewV4()
	requestID := req_uuid.String()
	return context.WithValue(context.Background(), RequestIDLabel, requestID)
}
