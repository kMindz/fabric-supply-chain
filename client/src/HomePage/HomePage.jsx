import React from 'react';
import {connect} from 'react-redux';
import {Link} from 'react-router-dom';

class HomePage extends React.Component {
  render() {
    const {user} = this.props;

    return (
      <div className="">
        <h1>Hi {user.name}!</h1>
        <div>
          <ul>
            <li><Link to='/products'>Products</Link></li>
            <li><Link to='/requests'>Requests</Link></li>
          </ul>
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  const {user} = state.authentication;
  return {
    user
  };
}

const connected = connect(mapStateToProps)(HomePage);
export {connected as HomePage};