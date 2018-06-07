from elastalert.alerts import Alerter, BasicMatchString
import boto3
from botocore.exceptions import ClientError
import os

class AwesomeNewAlerter(Alerter):
    def alert(self, matches):

        # Matches is a list of match dictionaries.
        # It contains more than one match when the alert has
        # the aggregation option set
        for match in matches:
            body_data = str(BasicMatchString(self.rule, match))
            SENDER = os.environ['SENDER'] #"ElastAlert <sw-devops@fossil.com>"
            RECIPIENT = os.environ['RECIPIENT'] #"sw-devops@fossil.com"

            AWS_REGION = os.environ['AWS_REGION'] #"us-east-1"

            SUBJECT = os.environ['SUBJECT'] #"[Misfit Store] WAF Alert"

            BODY_TEXT = body_data

            CHARSET = "UTF-8"

            client = boto3.client('ses',region_name=AWS_REGION)

            try:
                response = client.send_email(
                    Destination={
                        'ToAddresses': [
                            RECIPIENT,
                        ],
                    },
                    Message={
                        'Body': {
                            # 'Html': {
                            #     'Charset': CHARSET,
                            #     'Data': BODY_HTML,
                            # },
                            'Text': {
                                'Charset': CHARSET,
                                'Data': BODY_TEXT,
                            },
                        },
                        'Subject': {
                            'Charset': CHARSET,
                            'Data': SUBJECT,
                        },
                    },
                    Source=SENDER,
                )
            # Display an error if something goes wrong.   
            except ClientError as e:
                print(e.response['Error']['Message'])
            else:
                print("Email sent! Message ID:"),
                print(response['ResponseMetadata']['RequestId'])