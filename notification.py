"""
Function post_message sends Slack notification and response to thread with build status and
Trivy scanning results.
It looks for a scan_results.json file - this file is produced by the Scan Image step in the pipeline.
JH 
EM
"""
import os
import asyncio
import logging
from slack_sdk.web.async_client import AsyncWebClient
from slack_sdk.errors import SlackApiError

# Define the variables.
client = AsyncWebClient(token=os.environ['SLACK_BOT_TOKEN'])
channel_id = os.getenv('CHANNEL_ID')
filepath = "/drone/src/scan_results.json"
file_name = filepath
BUILD_LINK = os.getenv("BUILD_LINK")
BUILD_AUTHOR = os.getenv("BUILD_COMMIT_AUTHOR")
DRONE_BUILD_NUMBER = os.getenv("DRONE_BUILD_NUMBER")
BUILD_STATUS = os.getenv("BUILD_STATUS")
DRONE_BUILD_EVENT = os.getenv("DRONE_BUILD_EVENT")
SLACK_BOT = os.getenv("SLACK_BOT_TOKEN")
logging.basicConfig(level=logging.DEBUG)

# Determine vulnerabilities status.
# Open scan_results.json in read mode
with open('/drone/src/scan_results.json', 'r') as f:
    scan_results = f.read()
if "VulnerabilityID" in scan_results:
    VULNERABILITY = "failure"
elif "VulnerabilityID" not in scan_results:
    VULNERABILITY = "success"
else:
    VULNERABILITY = "unknown"
# Create a logger
logger = logging.getLogger(__name__)
VULNERABILITY = "failure"   # For testing purposes, remove this when testing completes.


async def post_message():
    logger.info("Running post_message info")
    try:
        if f"{VULNERABILITY}" == "failure":
            thread = await client.chat_postMessage(
                channel=channel_id,
                text="Hof build fall back message",
                blocks=[
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": "HOF Image Build Reports:"
                        }
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"{BUILD_LINK} "
                                    "Build Pipeline \n :x: :ambulance: :rotating-light-red: "
                                    ":fire_engine: \n Vulnerabilities found in Image!. \n" +
                                    f"Failure: Build {DRONE_BUILD_NUMBER} * (type: `{DRONE_BUILD_EVENT}`) \n" +
                                    f"Author: {BUILD_AUTHOR}"
                        },
                        "accessory": {
                            "type": "image",
                            # "image_url": "https://media.giphy.com/media/26tPjmWwr36k1OkYE/giphy.gif", # No Way! GIF
                            "image_url": "https://media3.giphy.com/media/26ybwvTX4DTkwst6U/200.gif?cid=6104955eoota1jfxhigqy3nb0a8e4mwpmxo36n6fjlblfnkh&rid=200.gif&ct=g",
                            # "image_url": "https://media0.giphy.com/media/l4FGlGcaAQbr7idTW/200.gif?cid=6104955er1czpbufdv159jkvrn2g4uoaol1l14b1vghyano1&rid=200.gif&ct=s", #Pixeled GIF
                            "alt_text": "Not good enough, try again"
                        }
                    },
                    {
                        "type": "section",
                        "fields": [
                            {
                                "type": "mrkdwn",
                                "text": "Hof build Scan found vulnerabilities.\n See below for more information"
                            }
                        ]
                    }
                ]
            )
            print(f"{DRONE_BUILD_EVENT}")
            response = await client.files_upload_v2(channel=channel_id,
                                                    file=file_name,
                                                    initial_comment="Vulnerability report for test :robot_face:",
                                                    title="Test File Upload",
                                                    thread_ts=thread["message"]["ts"])
        elif f"{VULNERABILITY}" == "success":
            thread_success = await client.chat_postMessage(
                channel=channel_id,
                text="Hof build success!",
                blocks=[
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": "HOF Image Scan Report:"
                        }
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"{BUILD_LINK} "
                                    "Build Pipeline> \n :hundred-points: :hammer_and_pick: :clap:"
                                    "\n WHOOO HOOO! Clean image built!. \n" +
                                    f"Success: Build {DRONE_BUILD_NUMBER} * (type: `{DRONE_BUILD_EVENT}`) \n" +
                                    f"Author: {BUILD_AUTHOR}"
                        },
                        "accessory": {
                            "type": "image",
                            "image_url": "https://media.giphy.com/media/xT5LMHxhOfscxPfIfm/giphy.gif",
                            "alt_text": "cute cat"
                        }
                    },
                    {
                        "type": "section",
                        "fields": [
                            {
                                "type": "mrkdwn",
                                "text": "Hof build completes."
                            }
                        ]
                    }
                ]
            )
            response = await client.reactions_add(channel=channel_id,
                                                  name="thumbsup",
                                                  timestamp=thread_success["message"]["ts"]
                                                  )
        else:
            thread_unknown = await client.chat_postMessage(
                channel=channel_id,
                text="Hof build status unknown"
                     "\nPlease investigate"
            )
            response = await client.reactions_add(channel=channel_id,
                                                  timestamp=thread_unknown["message"]["ts"],
                                                  name="rotating_light"
                                                  )

    except SlackApiError as e:
        assert e.response["ok"] is False
        assert e.response["error"]
        print(f"Got an error: {e.response['error']}")


asyncio.run(post_message())
