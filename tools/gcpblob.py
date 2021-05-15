#!/usr/bin/env python3
"""GCP blob storage interaction"""
import sys
from contextlib import contextmanager
from collections import namedtuple
from datetime import datetime
from pathlib import Path
import argparse
import base64
import json
import os
import random
import requests


import argparse


from google.auth._default import _CLOUD_SDK_CREDENTIALS_WARNING
from clicommon import run_and_exit

GENERAL_CONNECTION_EXEPTIONS = (
    requests.exceptions.Timeout,
    requests.exceptions.ConnectionError,
    ConnectionAbortedError,
    ConnectionRefusedError,
    ConnectionResetError)


class Error(Exception):
    """General, detected error condition"""


def get_object(
        token, bucket, objectname,
        check_response=True, isjson=True, generation=None):
    """Get the named object from the bucket

    If requested, get the generation from the metadata"""

    # blob storage names look heirarchial but arent. need to quote '/'
    objectname = requests.utils.quote(objectname, safe="")
    qs = f"https://storage.googleapis.com/storage/v1/b/{bucket}/o/{objectname}"

    if generation:
        resp = requests.get(qs, headers={"authorization": f"Bearer {token}"})
        if not resp and not check_response:
            return resp, 0
        generation = json_response(resp)["generation"]

    qs += "?alt=media"
    resp = requests.get(qs, headers={"authorization": f"Bearer {token}"})

    if not check_response:
        return resp, generation

    if isjson:
        return json_response(resp), generation
    return text_response(resp), generation


def get_blob(
        token, bucket, objectname, check_response=True, isjson=False):
    """convenience that doesn't do generation"""

    return get_object(
        token,  bucket, objectname, check_response=check_response, isjson=isjson)[0]


def post_blob(
        token, bucket, objectname, data,
        check_response=True, content_type="application/json", generation=None):
    """post the named object to the bucket"""

    qs = (
        f"https://storage.googleapis.com/upload/storage/v1/b/"
        f"{bucket}/o?uploadType=media&name={objectname}")

    if generation is not None:
        qs += f"&ifGenerationMatch={generation}"

    resp = requests.post(
        qs, data=data,
        headers={"authorization": f"Bearer {token}", "content-type": content_type})

    if not check_response:
        return resp

    return json_response(resp)


def cmd_post(args):
    """post a blob to gcp storage"""
    if args.token is None:
        raise Error("a token is required, try `gcloud auth print-access-token`")

    data = args.data
    if args.datafile == "-":
        data = sys.stdin.read()
    elif args.datafile:
        data = open(args.datafile).read()
    if data is None:
        raise Error("datafile or data must be provided")

    print(args.datafile)
    print(args.data)
    print(json.loads(data))
    print(args.token)
    print(args.bucket)
    print(args.objectname)

    resp = post_blob(
        args.token, args.bucket, args.objectname, data)
    print(json.dumps(resp, indent=2, sort_keys=True))


def json_response(resp):

    if not resp:
        raise Error(f"error: status {resp.status_code}")

    j = resp.json()
    if "error" not in j:
        return j

    if "message" in j["error"]:
        raise Error(j["error"]["message"])
    raise Error(f"error: status {resp.status_code}")


def text_response(resp):
    if resp:
        return resp.text
    raise Error(f"error: status {resp.status_code}")



def run(args=None):
    if args is None:
        args = sys.argv[1:]

    top = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    top.set_defaults(func=lambda a: print("see sub commands in help"))

    top.add_argument("-p", "--gcp-project", default="fetlar")
    top.add_argument("-b", "--bucket", default=None)
    top.add_argument(
        "-t", "--token", help="""
        provide bearer token for gcp api access (from, for example, gcloud auth
        print-access token)""")

    subcmd = top.add_subparsers(title="Available commands")

    p = subcmd.add_parser("post", help=cmd_post.__doc__)
    p.add_argument("objectname")
    p.add_argument("-f", "--datafile", help="file to read data from '-' to read from stdin")
    p.add_argument("-d", "--data", help="data string")
    p.set_defaults(func=cmd_post)

    args = top.parse_args()
    args.func(args)


if __name__ == "__main__":
    run_and_exit(run)
