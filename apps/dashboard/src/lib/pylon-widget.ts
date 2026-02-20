/* eslint-disable prefer-rest-params */
/* eslint-disable no-var */
/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const addPylonWidget = (appId: string) => {
  var e = window
  var t = document
  var n = function () {
    // @ts-expect-error ignore
    n.e(arguments)
  }
  // @ts-expect-error ignore
  n.q = []
  // @ts-expect-error ignore
  n.e = function (e) {
    // @ts-expect-error ignore
    n.q.push(e)
  }
  // @ts-expect-error ignore
  e.Pylon = n
  var r = function () {
    var e = t.createElement('script')
    e.setAttribute('type', 'text/javascript')
    e.setAttribute('async', 'true')
    e.setAttribute('src', `https://widget.usepylon.com/widget/${appId}`)
    var n = t.getElementsByTagName('script')[0]
    if (n.parentNode) n.parentNode.insertBefore(e, n)
  }
  if (t.readyState === 'complete') {
    r()
  } else if (e.addEventListener) {
    e.addEventListener('load', r, false)
  }
}
