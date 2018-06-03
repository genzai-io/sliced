/*
 * Copyright 2018 Movemedical, Inc. All rights reserved.
 *
 * Licensed under the terms of the LICENSE file distributed with this project.
 */

import {History} from 'history'
import {RouterStore} from './RouterStore'
import {APIStore} from "./APIStore"
import {LayoutStore} from "./LayoutStore";
import {STORE_API, STORE_ROUTER, STORE_LAYOUT} from 'app/constants'

export function createStores (history: History) {
  const router = new RouterStore(history)
  const api = new APIStore()
  const layout = new LayoutStore(router)
  return {
    [ STORE_ROUTER ]: router,
    [ STORE_API ]: api,
    [ STORE_LAYOUT ]: layout,
  }
}
