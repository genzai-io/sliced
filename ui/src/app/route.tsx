import * as React from "react";
import {Route, RouteComponentProps, RouteProps} from "react-router";

/**
 *
 */
export interface IPageRouteProps extends RouteProps {
  page: React.ComponentType<RouteComponentProps<any>> | React.ComponentType<any>
}

/**
 *
 */
export class PageRoute extends Route<IPageRouteProps> {
  render () {
    const { page, ...props } = this.props
    return React.createElement(page, { ...props })
  }
}