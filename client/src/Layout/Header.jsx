import React from 'react';
import {NavLink} from 'react-router-dom';
import {connect} from 'react-redux';

import {orgConstants} from '../_constants';

class Header extends React.Component {
  render() {
    const {user} = this.props;
    return (
      <nav className="navbar navbar-light navbar-expand">
        <div className="navbar-brand"><img src="./AltorosLogo_mini1.png" alt="logo"/></div>
        {user && <div className="container">
          <div>Hi, <b>{user.name}</b> from <i>{orgConstants[user.org]}</i></div>
          <ul className="nav navbar-nav pull-xs-right">

            <li className="nav-item">
              <NavLink exact to='/' className="nav-link">
                Home
              </NavLink>
            </li>

            <li className="nav-item">
              <NavLink to='/products' className="nav-link">
                Products
              </NavLink>
            </li>

            <li className="nav-item">
              <NavLink to='/requests' className="nav-link">
                Requests
              </NavLink>
            </li>

            <li className="nav-item">
              <NavLink to='login' className="nav-link">
                Logout
              </NavLink>
            </li>
          </ul>
        </div>}
      </nav>
    );
  }
}

function mapStateToProps(state) {
  const {user} = state.authentication;
  return {
    user
  };
}

const connected = connect(mapStateToProps)(Header);
export {connected as Header};