import * as React from 'react'
import {inject, observer} from 'mobx-react'
import {STORES} from 'app/constants'
import {APIStore, LayoutStore, NavModule, RouterStore} from 'app/stores'

import {Icon} from '@blueprintjs/core'
import {IconName, IconNames} from '@blueprintjs/icons'

import './style.css'

@inject(...STORES)
@observer
export class NavBar extends React.Component<{
  layout?: LayoutStore
  router?: RouterStore
}, {}> {
  private onOpenNav = () => {
    console.log('toggleNav')
    this.props.layout.toggleNav()
  }

  render () {
    return (
      <nav className='pt-navbar navbar'>
        {/*<div style='margin: 0 auto width: 480px'>*/}
        <div>
          <div className='pt-navbar-group pt-align-left nav-drawer-icon'>
            <table cellPadding={0} cellSpacing={0}>
              <tbody>
              <tr>
                <td className='nav-open-container'>
                  <button type='button' onClick={this.onOpenNav}
                          className='pt-button pt-minimal pt-icon-menu'/>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
          <div className='pt-navbar-group pt-align-left'>
            {this.props.layout.title}
          </div>
          <div className='pt-navbar-group pt-align-right'>
            <button className='pt-button pt-minimal pt-icon-user'></button>
            <button className='pt-button pt-minimal pt-icon-notifications'
                    onClick={() => this.props.router.goToAlerts()}></button>
            <button className='pt-button pt-minimal pt-icon-cog'
                    onClick={() => this.props.router.goToSettings()}></button>
          </div>
        </div>
      </nav>
    )
  }
}

// const logo = require('./mm-logo-m-white.svg')
// const logo = require('./logo.png')

@inject(...STORES)
@observer
export class Nav extends React.Component<{
  layout?: LayoutStore
  router?: RouterStore
  browser?: APIStore
}, {}> {
  componentWillMount () {
  }

  render () {
    const { layout, router } = this.props
    const selected = layout.navModule

    return (
      <div>
        <div className='brand-container'>
          <div style={{ textAlign: 'center' }}>
            <svg width='31px' height='31px' viewBox='0 0 31 31' version='1.1' onMouseEnter={() => {
            }} onMouseLeave={() => {
            }}>
              <g id='Homepage' stroke='none' strokeWidth='1' fill='none' fillRule='evenodd'>
                <g id='Home-1440' transform='translate(-1076.000000, -318.000000)' fill={layout.logoColor}>
                  <g id='Group-4' transform='translate(373.000000, 318.000000)'>
                    <path
                      d='M703,6.05821785 L703,30.9999997 L715.417525,31 C715.417525,31 726.215385,31 725.996723,19.2343048 C725.996723,8.96448357 725.996723,10.0168671 725.996723,10.0168671 L718.703306,19.2343051 L711.520727,10.0498779 L711.520727,25.7775348 L708.460916,25.7775348 L708.460916,6 L703,6.05821785 Z M734,-1.38555833e-13 L733.999999,16 L728.823302,16 L728.823303,5.12111165 L725.858375,5.12111165 L718.744144,14.1668526 L712,5.12111169 C712,5.12111169 713.621991,-5.89789039e-08 718.744144,-1.38555833e-13 C727.324738,9.88006621e-08 734,-1.38555833e-13 734,-1.38555833e-13 Z'
                      id='Combined-Shape'></path>
                  </g>
                </g>
              </g>
            </svg>
          </div>
          <div className='brand-title'>Move Medical</div>
          <div className='brand-text'>Cloud Platform</div>
        </div>
        <ul className='nav-menu pt-list-unstyled'>
          <Item
            ui={layout}
            router={router}
            name={'Dashboard'}
            module={NavModule.DASHBOARD}
            icon={IconNames.DASHBOARD}
            selected={selected === NavModule.DASHBOARD}/>

          <Item
            ui={layout}
            router={router}
            name='Cluster'
            module={NavModule.CLUSTER}
            icon={IconNames.CLOUD}
            selected={selected === NavModule.CLUSTER}/>

          <Item
            ui={layout}
            router={router}
            name={'Actions'}
            module={NavModule.ACTIONS}
            icon={IconNames.JOIN_TABLE}
            selected={selected === NavModule.ACTIONS}/>

          <Item
            ui={layout}
            router={router}
            name={'Logs'}
            module={NavModule.QUEUES}
            icon={IconNames.DATABASE}
            selected={selected === NavModule.QUEUES}/>

          <Item
            ui={layout}
            router={router}
            name={'Alerts'}
            module={NavModule.ALERTS}
            icon={IconNames.NOTIFICATIONS}
            selected={selected === NavModule.ALERTS}/>

          <Item
            ui={layout}
            router={router}
            name={'Settings'}
            module={NavModule.SETTINGS}
            icon={IconNames.SETTINGS}
            selected={selected === NavModule.SETTINGS}/>
        </ul>
      </div>
    )
  }
}

interface IItemProps {
  ui: LayoutStore
  router: RouterStore
  name: string
  module: NavModule
  icon: IconName | JSX.Element | false | null | undefined
  selected: boolean
}

@observer
class Item extends React.Component<IItemProps, {}> {
  private onClick = () => {
    if (!this.props.selected) {
      this.props.router.push(LayoutStore.navRoute(this.props.module))
    }
  }

  render () {
    const { name, selected } = this.props
    return (
      <li className={selected ? 'selected' : ''} onClick={this.onClick}>
        <Icon icon={this.props.icon} iconSize={20}/> <span style={{ paddingLeft: '5px', lineHeight: '20px' }}>{name}</span>
      </li>
    )
  }
}
