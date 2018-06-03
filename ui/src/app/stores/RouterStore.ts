import {History} from 'history'
import {RouterStore as BaseRouterStore, syncHistoryWithStore} from 'mobx-react-router'
import {ROUTE_ALERTS, ROUTE_AUTH, ROUTE_HOME, ROUTE_SETTINGS} from "app/constants";

export class RouterStore extends BaseRouterStore {
  constructor (history?: History) {
    super()
    if (history) {
      this.history = syncHistoryWithStore(history, this)
    }
  }

  goToHome () {
    this.push(ROUTE_HOME)
  }

  goToAuth () {
    this.push(ROUTE_AUTH)
  }

  goToSettings () {
    this.push(ROUTE_SETTINGS)
  }

  goToAlerts () {
    this.push(ROUTE_ALERTS)
  }
}

export default RouterStore
