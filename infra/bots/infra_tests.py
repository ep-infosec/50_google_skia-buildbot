#!/usr/bin/env python
#
# Copyright 2016 Google Inc.
#
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.


"""Run all infrastructure-related tests."""


from __future__ import print_function
import os
import subprocess
import sys


INFRA_BOTS_DIR = os.path.dirname(os.path.realpath(__file__))
SKIA_DIR = os.path.abspath(os.path.join(INFRA_BOTS_DIR, os.pardir, os.pardir))


def test(cmd, cwd):
  try:
    subprocess.check_output(cmd, cwd=cwd, stderr=subprocess.STDOUT,
                            encoding='utf-8')
  except subprocess.CalledProcessError as e:
    return e.output


def gen_tasks_test(train):
  cmd = ['bazelisk', 'run', '//:go', '--', 'run', 'gen_tasks.go']
  if not train:
    cmd.append('--test')
  try:
    output = test(cmd, INFRA_BOTS_DIR)
  except OSError:
    return ('Failed to run "%s"; do you have Go installed on your machine?'
            % ' '.join(cmd))
  if output and 'cannot find package "go.skia.org/infra' in output:
    return ('Failed to run gen_tasks.go:\n\n%s\nMaybe you need to run:\n\n'
            '$ go get -u go.skia.org/infra/...' % output)
  return output


def main():
  train = False
  if '--train' in sys.argv:
    train = True

  tests = (
      gen_tasks_test,
  )
  errs = []
  for t in tests:
    err = t(train)
    if err:
      errs.append(err)

  if len(errs) > 0:
    print('Test failures:\n', file=sys.stderr)
    for err in errs:
      print('==============================', file=sys.stderr)
      print(err, file=sys.stderr)
      print('==============================', file=sys.stderr)
    sys.exit(1)

  if train:
    print('Trained tests successfully.')
  else:
    print('All tests passed!')


if __name__ == '__main__':
  main()
