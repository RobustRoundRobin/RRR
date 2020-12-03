#!/usr/bin/env python3
"""Support for creating and issuing raw transactions using local private keys

Cant do this directly with truffle afaict, (though we could probably have a js
implementation). However, we *do* leverage truffle's support for compiling
contracts and their abis.

There are web3 things we could replace this with

* https://gist.github.com/tomconte/6ce22128b15ba36bb3d7585d5180fba0
* https://www.npmjs.com/package/web3js-raw

"""
import sys
import time
import argparse
import json
import subprocess as sp
from pathlib import Path
import textwrap
import traceback
import rlp, rlp.sedes
from sha3 import keccak_256

import coincurve

from eth_keys.datatypes import PrivateKey

DEFAULT_GAS_PRICE = 0
DEFAULT_GAS_LIMIT = 500000000

class Error(Exception):
    """General error conditions that don't benefit from a traceback"""


def big_endian_int(x):
    """big endian to integer conversion using rlp.sedes"""
    return rlp.sedes.big_endian_int.deserialize(x.lstrip(b'\x00'))

def keccak(message):
    """keccak convenience for sha3.keccak_256"""
    return keccak_256(message).digest()

# pylint: disable=too-many-arguments
def encoderaw(nonce, gasprice, gaslimit, address, data, chainid, key=None, verbose=False):
    """Encode and optionaly sign a transaction

    Given all the necessary parameters. (Note: assumes the chain has EIP-155
    enabled and includes all 9 elements)

    Args:
        nonce, gasprice, gas:
            corresponding integers
        address:
            hex address
        data:
            hex data
        key:
            key (bytes)
    """

    if data.startswith("0x"):
        data = data[2:]
    message = rlp.encode([
        nonce, gasprice, gaslimit, int(address, 16) if address else b'',
        0, bytes.fromhex(data), chainid, b'', b''])

    if not key:
        return message, None, None, None

    key = coincurve.PrivateKey(key)

    pub_key = key.public_key.format(compressed=False)[1:] # remove the compressed indicator byte
    addr = keccak_256(pub_key).digest()[-20:]
    sig = key.sign_recoverable(keccak(message), hasher=None)
    r = big_endian_int(sig[0:32])
    s = big_endian_int(sig[32:64])

    # v = sig[64] + 27 # is how eth_keys does this but private chain (quorum)
    # is different. The following assumes EIP 155 is enabled for the chain at
    # block zero.
    #
    # https://bitcoin.stackexchange.com/questions/38351/ecdsa-v-r-s-what-is-v
    # https://docs.goquorum.com/en/latest/Getting%20Started/running/ (comments
    # re EIP 155), and EIP 155 itself
    if r % 2 == 1:
        v = chainid * 2 + 36
    else:
        v = chainid * 2 + 35

    if verbose:
        print(f"signer-addr: {addr.hex()}, nonce={nonce}, chain-id={chainid}, v={v}, r={hex(r)}, s={hex(s)}")

    signed = rlp.encode([
        nonce, gasprice, gaslimit, int(address, 16) if address else b'',
        0, bytes.fromhex(data), v, r, s])

    # pubver = coincurve.PublicKey.from_signature_and_message(sig, keccak_256(message).digest(), hasher=None)

    return message, signed, addr, (v, r, s)


def cmd_encode(args):
    """RLP Encode and optionaly sign a transaction. Optionaly executing it.

    If a contract address is provided, the abi from the contract file (truffle
    json) is used to create a contract function call transaction. Extra
    positional cli arguments are encoded as call parameters.

    Othwerwise a transaction is created to deploy the bytecode from the
    contract file.
    """

    geth_cmd = geth_exec_cmd(args)

    c = json.load(args.contract)

    if args.to is not None:
        abi = c["abi"]
        js = textwrap.dedent(f"""
        var abi = new web3.eth.contract(%s);
        abi.at("{args.address}").{args.method}.getData({args.method_args});
        """) % abi  # formating trick to sort out left alignent

        cmd = geth_cmd + [js]
        data = jserr(args, sp.check_output(cmd).decode().strip(" \"\n"))
        if data.startswith("0x"):
            data = data[2:]
    else:
        data = c["bytecode"]

    if not args.priv and not args.hex_priv:
        message, *_ = encoderaw(
            args.nonce, args.gasprice, args.gaslimit, args.to, data, args.chainid,
            verbose=args.verbose)
        print(message.hex())
        return 0

    if args.priv:
        key = args.priv.read()
    else:
        key = bytearray.fromhex(args.hex_priv.read())


    message, signed, *_ = encoderaw(
            args.nonce, args.gasprice, args.gaslimit, args.to, data, args.chainid,
            key=key, verbose=args.verbose)

    js = f'eth.sendRawTransaction("0x{signed.hex()}")'
    formats = args.format.split(",")
    if "hex" in formats:
        print("0x{signed.hex()}")
    if "js" in formats:
        print(js)
    return 0


def geth_exec_cmd(args):
    """geth attach command from args"""
    return [args.geth_attach, args.geth_ipc, "--exec"]


def print_exc():
    """Compact representation of current exception

    Single line tracebacks are individually a little confusing but prevent
    other useful output from being obscured"""

    exc_info = sys.exc_info()
    trace = [
        f"{Path(fn).name}[{ln}].{fun}:\"{txt}\""
        for (fn, ln, fun, txt) in traceback.extract_tb(exc_info[2])
    ] + [f"{exc_info[0].__name__}:{exc_info[1]}"]

    print("->".join(trace), file=sys.stderr)


def jserr(out, js=None):
    """raise Error if output indicates error executing js"""

    if "Error" in out or "error" in out:
        if js is not None:
            out = '\n\n'.join([js, out])
        raise Error(out)
    return out


def run(args=None):
    """Main entry point"""
    if args is None:
        args = sys.argv[1:]

    top = argparse.ArgumentParser(description=__doc__)
    top.add_argument('-v', '--verbose', action='store_true')

    cmds = top.add_subparsers(title="Commands")
    p = cmds.add_parser("encode", help=cmd_encode.__doc__)
    p.set_defaults(func=cmd_encode)

    p.add_argument("contract", type=argparse.FileType('r'),
        help="""
        truffle compatible <contract>.json"""
    )
    p.add_argument("--format", default="js", help="comman seperated list of output formats. any or all of: js,hex")
    p.add_argument("--geth-attach", default="geth attach")
    p.add_argument("--geth-ipc", default="geth.ipc")

    p.add_argument("--nonce", "-n", default=0, type=int)
    p.add_argument("--gasprice", default=DEFAULT_GAS_PRICE)
    p.add_argument("--gaslimit", "-g", default=DEFAULT_GAS_LIMIT)

    p.add_argument("--to", help="""
    transaction `to' address, typically a contract address. to deploy the
    contract *omit* this""")

    p.add_argument("--chainid", default=99, type=int)

    p.add_argument("--method", "-m", help="contract method function")
    p.add_argument("--method-args", "-a", default="")

    p.add_argument("-k", "--priv", type=argparse.FileType('rb'),
        help="""
        Sign the raw encoded transaction with the binary encoded key in this file"""
    )
    p.add_argument("-x", "--hex-priv", type=argparse.FileType('r'),
        help="""
        Sign the raw encoded transaction with the hex encoded key in this file"""
    )

    args = top.parse_args(args)
    args.func(args)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(run())
    except Error as e:
        print(e, file=sys.stderr)
    except Exception:
        print_exc()
    sys.exit(-1)
