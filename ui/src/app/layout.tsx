import * as React from 'react'
// import * as style from './style.css'
import {inject, observer} from 'mobx-react'
import {
  ROUTE_ACTIONS,
  ROUTE_ALERTS,
  ROUTE_AUTH,
  ROUTE_CLUSTER,
  ROUTE_DASHBOARD,
  ROUTE_HOME,
  ROUTE_LOGS,
  ROUTE_SETTINGS,
  STORE_LAYOUT,
  STORE_ROUTER,
  STORES
} from 'app/constants'
import {Hotkey, Hotkeys, Position, Toaster} from '@blueprintjs/core'
import {RouteComponentProps, Switch} from 'react-router'

// Inject style.css
import './style.css'

import {Auth} from 'app/components/Auth'
import {Home} from 'app/components/Home'
import {Nav, NavBar} from 'app/components/Nav'
import {Dashboard} from 'app/components/Dashboard/Home'
import {Cluster} from 'app/components/Cluster/Home'
import {Actions} from 'app/components/Actions/Home'
import {Logs} from 'app/components/Logs/Home'
import {Alerts} from 'app/components/Alerts/Home'
import {Settings} from 'app/components/Settings'
import {LayoutStore, RouterStore} from 'app/stores'
import {PageRoute} from 'app/route'

const AppToaster = Toaster.create({
  position: Position.TOP,
})

export interface AppProps extends RouteComponentProps<any> {
  [ STORE_LAYOUT ]: LayoutStore
  [ STORE_ROUTER ]: RouterStore
}

@inject(...STORES)
@observer
export class Layout extends React.Component<AppProps, {}> {
  layout: LayoutStore = this.props[ STORE_LAYOUT ]

  showToast = () => {
    // create toasts in response to interactions.
    // in most cases, it's enough to simply create and forget (thanks to timeout).
    AppToaster.show({ message: 'Toasted.' })
  }

  private onWindowResize = () => {
    this.layout.windowResized(window.innerWidth, window.innerHeight)
  }

  // private get router (): RouterStore {
  //   return this.props.router
  // }
  private onMainMouseDown = () => {
    if (this.layout.isCollapse && this.layout.navClass == 'nav-open') {
      this.layout.toggleNav()
    }
  }

  public renderHotkeys () {
    return (
      <Hotkeys tabIndex={null}>
        <Hotkey global={true} label='Focus the piano' combo='shift + p'/>
      </Hotkeys>
    )
  }

  componentDidMount () {
    window.addEventListener('resize', this.onWindowResize)
    this.layout.windowResized(window.innerWidth, window.innerHeight)
  }

  componentWillUnmount (): void {
    window.removeEventListener('resize', this.onWindowResize)
  }

  render () {
    // const { children } = this.props

    return (
      <div className='container' style={{ width: '100%' }}>
        {/*<header></header>*/}
        <nav className={this.layout.navClass}>
          {/*<nav className='nav-closed'>*/}
          <div className='nav-container'>
            <Nav/>
          </div>
        </nav>
        <main className='main-container' onMouseDown={this.onMainMouseDown}>
          <NavBar/>

          <Switch>
            <PageRoute path={ROUTE_AUTH} page={Auth}/>
            <PageRoute path={ROUTE_DASHBOARD} page={Dashboard}/>
            <PageRoute path={ROUTE_CLUSTER} page={Cluster}/>
            <PageRoute path={ROUTE_ACTIONS} page={Actions}/>
            <PageRoute path={ROUTE_LOGS} page={Logs}/>
            <PageRoute path={ROUTE_ALERTS} page={Alerts}/>
            <PageRoute path={ROUTE_SETTINGS} page={Settings}/>

            <PageRoute path={ROUTE_HOME} page={Home}/>
          </Switch>
        </main>
        {/*<aside>Related links</aside>*/}
        {/*<footer></footer>*/}
      </div>
    )
  }
}
