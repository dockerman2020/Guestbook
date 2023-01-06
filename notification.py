"""
Function is to send Slack notification and response to thread with build status and
Trivy scanning results.
"""
import os
import asyncio
import logging
from slack_sdk.web.async_client import AsyncWebClient
from slack_sdk.errors import SlackApiError

client = AsyncWebClient(token=os.environ['SLACK_BOT_TOKEN'])
channel_id = os.getenv('channel_id')
filepath = "/drone/src/scan_results.json"
file_name = filepath
BUILD_LINK = os.getenv("BUILD_LINK")
BUILD_AUTHOR = os.getenv("BUILD_AUTHOR")
DRONE_BUILD_NUMBER = os.getenv("DRONE_BUILD_NUMBER")
BUILD_STATUS = os.getenv("BUILD_STATUS")
DRONE_BUILD_EVENT = os.getenv("DRONE_BUILD_EVENT")
SLACK_BOT = os.getenv("SLACK_BOT_TOKEN")
logging.basicConfig(level=logging.DEBUG)
# Create a logger
logger = logging.getLogger(__name__)

# print(BUILD_AUTHOR, BUILD_LINK, BUILD_STATUS, DRONE_BUILD_EVENT, DRONE_BUILD_NUMBER, SLACK_BOT)

async def post_message():
    logger.debug("Running post_message debug")
    try:
        if f"{BUILD_STATUS}" == "failure":
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
                                    "Build Pipeline> \n :x: :ambulance: :rotating-light-red: "
                                    ":fire_engine: \n Vulnerabilities found in Image!. \n" +
                                    f"Failure: Build {DRONE_BUILD_NUMBER} * (type: `{DRONE_BUILD_EVENT}`) \n" +
                                    f"Author: {BUILD_AUTHOR}"
                        },
                        "accessory": {
                            "type": "image",
                            "image_url": "https://media.giphy.com/media/26tPjmWwr36k1OkYE/giphy.gif",
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
            print(f"{DRONE_BUILD_EVENT}")
            response = await client.files_upload_v2(channel=channel_id,
                                                    file=file_name,
                                                    initial_comment="Vulnerability report for test :robot_face:",
                                                    title="Test File Upload",
                                                    thread_ts=thread["message"]["ts"])
        elif f"{BUILD_STATUS}" == "success":
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
