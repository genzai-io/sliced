import {action, computed, observable} from 'mobx';
import {
  ROUTE_ACTIONS,
  ROUTE_ALERTS,
  ROUTE_CLUSTER,
  ROUTE_DASHBOARD,
  ROUTE_HOME,
  ROUTE_LOGS,
  ROUTE_SETTINGS
} from "app/constants";
import RouterStore from "app/stores/RouterStore";
import {Action, Location} from "history";
import {ReactElement} from "react";
import {Colors, IToaster, IToastProps, Position, Toaster} from "@blueprintjs/core";

const MAX_WIDTH = 768
const CSS_NAV_CLOSE = 'nav-close'
const CSS_NAV_CLOSED = 'nav-closed'
const CSS_NAV_OPEN = 'nav-open'
const CSS_DARK = 'pt-dark'
const CSS_FULLSCREEN = 'fullscreen'
const CSS_FULLSCREEN_DARK = `${CSS_FULLSCREEN} ${CSS_DARK}`

export class Notification {
  constructor (public code: number, public message: string) {
  }
}

export enum NavModule {
  NONE = 0,
  DASHBOARD = 1,
  CLUSTER = 2,
  ACTIONS = 3,
  QUEUES = 4,
  ALERTS = 5,
  SETTINGS = 6
}

export class LayoutStore {
  @observable public fullScreen: boolean
  @observable public dark: boolean
  @observable public navClass: string
  @observable public width: number
  @observable public height: number
  @observable public navModule: NavModule
  @observable public title: ReactElement<any>
  private readonly toaster: IToaster
  private readonly body: HTMLElement
  private readonly router: RouterStore

  constructor (router: RouterStore) {
    this.router = router
    this.toaster = Toaster.create({
      position: Position.TOP,
    })
    this.navClass = 'nav-closed'
    this.width = window.innerWidth
    this.height = window.innerHeight
    this.body = document.body
    this.body.className = ''
    this.navModule = NavModule.NONE

    this.router.history.listen(this.onHistory.bind(this))
    this.selectNavModule(this.router.location.pathname)

    if (this.canCollapseNav) {
      this.toggleNav()
    }

    this.theme(false)
  }

  @computed get isCollapse () {
    return this.width < MAX_WIDTH
  }

  @computed get logoColor () {
    return this.dark ? Colors.WHITE : Colors.GRAY1
  }

  @computed get canCollapseNav () {
    return this.isCollapse && this.navClass === 'nav-open'
  }

  static navRoute (module: NavModule): string {
    switch (module) {
      case NavModule.DASHBOARD:
        return ROUTE_DASHBOARD
      case NavModule.CLUSTER:
        return ROUTE_CLUSTER
      case NavModule.ACTIONS:
        return ROUTE_ACTIONS
      case NavModule.QUEUES:
        return ROUTE_LOGS
      case NavModule.ALERTS:
        return ROUTE_ALERTS
      case NavModule.SETTINGS:
        return ROUTE_SETTINGS
    }
    return ROUTE_HOME
  }

  @action toast (props: IToastProps) {
    this.toaster.show(props)
  }

  @action onHistory (location: Location, action: Action) {
    const path = location.pathname.toLowerCase()

    this.selectNavModule(path)

    if (this.canCollapseNav) {
      this.toggleNav()
    }
  }

  @action selectNavModule (path: string) {
    if (path.startsWith(ROUTE_DASHBOARD)) {
      this.navModule = NavModule.DASHBOARD
    } else if (path.startsWith(ROUTE_CLUSTER)) {
      this.navModule = NavModule.CLUSTER
    } else if (path.startsWith(ROUTE_ACTIONS)) {
      this.navModule = NavModule.ACTIONS
    } else if (path.startsWith(ROUTE_LOGS)) {
      this.navModule = NavModule.QUEUES
    } else if (path.startsWith(ROUTE_ALERTS)) {
      this.navModule = NavModule.ALERTS
    } else if (path.startsWith(ROUTE_SETTINGS)) {
      this.navModule = NavModule.SETTINGS
    } else {
      this.navModule = NavModule.NONE
    }
  }

  @action theme (dark: boolean) {
    if (this.dark != dark) {
      this.body.className = dark
        ? this.fullScreen
          ? CSS_FULLSCREEN_DARK : CSS_DARK
        : this.fullScreen
          ? CSS_FULLSCREEN : ''
      this.dark = dark
    }
  }

  @action setFullScreen (fullScreen: boolean) {
    if (this.fullScreen != fullScreen) {
      if (this.body.className.indexOf(CSS_DARK) > -1) {
        this.body.className = fullScreen ? CSS_FULLSCREEN_DARK : CSS_DARK
      } else {
        this.body.className = fullScreen ? CSS_FULLSCREEN : ''
      }
      this.fullScreen = fullScreen
    }
  }

  @action windowResized (width: number, height: number) {
    this.width = width
    this.height = height
    this.navClass = CSS_NAV_CLOSE
  }

  @action toggleNav () {
    if (this.isCollapse) {
      this.navClass =
        this.navClass == CSS_NAV_CLOSE || this.navClass == CSS_NAV_CLOSED
          ? CSS_NAV_OPEN
          : CSS_NAV_CLOSE
    } else {
      this.navClass = CSS_NAV_CLOSED
    }
  }
}