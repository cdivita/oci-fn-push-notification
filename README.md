# fn-push-notifications
A [Fn Project](https://fnproject.io/) function for sending mobile push notifications through third party integrations.

Currently, push notifications can be sent to Androind and iOS devices using [Firebase Cloud Messaging (FCM)](https://firebase.google.com/docs/cloud-messaging).

# Requirements
For sending push notifications, a [Firebase Admin SDK service account](https://firebase.google.com/docs/admin/setup#set-up-project-and-service-account) is required.

Of course, also an OCI tenant where deploy the function is required.

# Constraints
Currently the most recent [Firebase Admin Go SDK](https://pkg.go.dev/firebase.google.com/go/v4) supported version is [`4.9.0`](https://pkg.go.dev/firebase.google.com/go/v4@v4.9.0), because [Go Fn FDK](https://github.com/fnproject/docs/tree/master/fdks/fdk-go) doesn't support `go` versions `>1.15`.

# Configuration
The function supports the following [configuration variables](https://github.com/fnproject/docs/blob/master/fn/develop/configs.md):

|Variable                                    | Description                                                                                                    | Required                                                       | Notes                                                                                                                      |
|:-                                          |:-                                                                                                              |:-                                                              |:-                                                                                                                          |
| `GOOGLE_APPLICATION_CREDENTIALS`           | The Firebase Admin SDK service account private key file, in JSON format. The variable should be base64 encoded | If `GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID` is not specified | Configuration variables are stored in plain text, please consider using `GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID` instead |
| `GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID` | The OCID of the secret that contains the Firebase Admin SDK service account private key file, in JSON format   | If `GOOGLE_APPLICATION_CREDENTIALS` is not specified           |

# The function invocation payload
Most of the parameters required for sending a SMS can be set through the config variables, therefore just few information are required to successfully invoke the function..

The function payload is a JSON with the following attributes:

|Name             | Description                                             | Notes                                              |
|:-               |:-                                                       |:-                                                  |
| `recipients`    | The registration tokens of push notification recipients |                                                    |
| `data`          | A custom set of key-value pairs                         | Only strings are allowed, both for keys and values |
| `message`       | The message to push to recipients                       |                                                    |
| `message.title` | The title of the message                                |                                                    |
| `message.body`  | The body of the message                                 |                                                    |

An example of function's payload is the following:
```json
{
    "recipients": [
        "00000000"
    ],
    "data": {
        "description": "Use the data element for sending custom key-value pairs",
        "note": "Only strings are allowed, both for keys and values"
    },
    "message": {
        "title": "Hello",
        "body": "A message for you !!!"
    }
}'
```

# Running the function

## Prerequisites
1. Configure [your tenant](https://docs.cloud.oracle.com/en-us/iaas/Content/Functions/Tasks/functionsconfiguringtenancies.htm) and your [development environment](https://docs.cloud.oracle.com/en-us/iaas/Content/Functions/Tasks/functionsconfiguringclient.htm) to use Oracle Functions
2. A [Firebase Admin SDK service account](https://firebase.google.com/docs/admin/setup#set-up-project-and-service-account)

## Deployment
For deploying the function:
1. Clone the oci-fn-push-notification repo
   - `git clone https://github.com/cdivita/oci-fn-push-notification.git`
2. Create the application in OCI Functions
   - `fn create app <app-name> --annotation oracle.com/oci/subnetIds='["<subnet-ocid>"]'`
3. Deploy the function
   - `fn -v deploy --app <app-name> --no-bump`
4. Configure the function
   - `fn config function <app-name> fn-push-notification GOOGLE_APPLICATION_CREDENTIALS <Firebase Admin SDK service account private key>`

If you stored the Google Application Credentials [OCI Vault](https://docs.cloud.oracle.com/en-us/iaas/Content/KeyManagement/Tasks/managingsecrets.htm), the configuration activities are slightly different:
1. Use `GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID` rather than `GOOGLE_APPLICATION_CREDENTIALS` as configuration variable
   - `fn config function <app-name> fn-push-notification GOOGLE_APPLICATION_CREDENTIALS_SECRET_ID <The OCID of your Twilio Auth Token secret>`
2. Create a dynamic group that includes your function resources. A matching rule that can be used is: `all {resource.type = 'fnfunc'}`
3. Create a policy for such dynamic group that allows the access to keys resources
   -  `allow dynamic-group <dynamic-group-name> to read secret-bundles in compartment <keys-compartment>`