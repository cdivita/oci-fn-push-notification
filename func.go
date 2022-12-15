package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"fnpush/model"
	"fnpush/oci"
	"fnpush/providers/fcm"
	"io"
	"math"
	"net/http"
	"strconv"

	fdk "github.com/fnproject/fdk-go"
)

const headerContentType = "Content-Type"
const contentTypeJson = "application/json"

const ConfigGoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
const ConfigGoogleApplicationCredentialsSecretId = "GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID"
const ConfigMaxRecipientsCount = "MAX_RECIPIENTS_COUNT"

func main() {
	fdk.Handle(fdk.HandlerFunc(notify))
}

func assertContentType(fn fdk.Context, input io.Reader, output io.Writer) error {

	contentType := fn.ContentType()
	if contentType != contentTypeJson {
		return fmt.Errorf("unsupported Content Type: %s", contentType)
	}

	return nil
}

func createOciClient(ctx context.Context, fn fdk.Context) (oci.OciClient, error) {

	profile := fn.Config()["OCI_PROFILE"]

	// If no profile is specified, Instance Principal Authentication is assumed
	if profile == "" {
		return oci.NewClient()
	}

	path := fn.Config()["OCI_CONFIG_LOCATION"]

	return oci.NewClientWithConfiguration(&path, &profile)
}

func createFirebaseClient(ctx context.Context, fn fdk.Context) (fcm.FcmClient, error) {

	var credentials []byte
	var err error

	credentialsSecretId := fn.Config()[ConfigGoogleApplicationCredentialsSecretId]

	if credentialsSecretId != "" {

		var ociClient oci.OciClient
		ociClient, err = createOciClient(ctx, fn)

		if err != nil {
			return nil, err
		}

		fmt.Printf("Loading Firebase Admin SDK credentials from OCI secret %s\n", credentialsSecretId)

		credentials, err = ociClient.GetSecret(ctx, credentialsSecretId)
	} else {

		credentialsBase64 := fn.Config()[ConfigGoogleApplicationCredentials]
		credentials, err = base64.StdEncoding.DecodeString(credentialsBase64)

		fmt.Println("Loading Firebase Admin SDK credentials from GOOGLE_APPLICATION_CREDENTIALS (please store it within an OCI secret and provide its OCID through GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID)")
	}

	if err != nil {
		return nil, err
	}

	return fcm.NewClient(credentials)
}

func notify(ctx context.Context, input io.Reader, output io.Writer) {

	fn := fdk.GetContext(ctx)

	if err := assertContentType(fn, input, output); err != nil {
		replyWithError(fn, output, http.StatusUnsupportedMediaType, err)
		return
	}

	request := new(model.PushRequest)
	err := json.NewDecoder(input).Decode(request)

	if err != nil {
		replyWithError(fn, output, http.StatusUnprocessableEntity, err)
		return
	}

	maxRecipients, _ := strconv.Atoi(fn.Config()[ConfigMaxRecipientsCount])
	maxRecipients = int(math.Min(float64(maxRecipients), fcm.MaxAllowedRecipients))

	if maxRecipients > 0 && len(request.Recipients) > maxRecipients {
		replyWithError(fn, output, http.StatusForbidden, fmt.Errorf("too many recipients (allowed: %d)", maxRecipients))
		return
	}

	fcmClient, err := createFirebaseClient(ctx, fn)

	if err != nil {
		replyWithError(fn, output, http.StatusInternalServerError, err)
		return
	}

	response, err := fcmClient.Push(request)

	if err != nil {
		replyWithError(fn, output, http.StatusInternalServerError, err)
	} else {
		reply(fn, output, http.StatusOK, response)
	}
}

func replyWithError(fn fdk.Context, output io.Writer, status int, e error) {

	fmt.Printf("%s (status: %d)\n", e.Error(), status)

	reply(fn, output, status, model.Error{
		Status:  status,
		Message: e.Error(),
	})
}

func reply(fn fdk.Context, output io.Writer, status int, object interface{}) {

	fdk.WriteStatus(output, status)
	fdk.SetHeader(output, headerContentType, contentTypeJson)

	json.NewEncoder(output).Encode(&object)
}
