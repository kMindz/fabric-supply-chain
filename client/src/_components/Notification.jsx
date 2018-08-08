import React from 'react';
import {connect} from 'react-redux';
import './Notification.css';
import {alertActions} from '../_actions/alert.actions';

class Notification extends React.Component {
  constructor(props) {
    super(props);

    this.cancel = this.cancel.bind(this);
  }
  render() {
    const {alert} = this.props;
    return (
      <div className={'notification ' + (alert.message ? 'fadein': 'fadeout invisible')}>
        {alert.message &&
        <div className={`d-flex align-items-between  alert ${alert.type || ''}`}>
          <span>{alert.message}</span>
          <span aria-hidden="true" onClick={this.cancel} className="close-btn">&times;</span>
        </div>}
      </div>
    );
  }

  cancel() {
    this.props.dispatch(alertActions.clear());
  }
}

function mapStateToProps(state) {
  const {alert} = state;
  return {
    alert
  };
}

const connected = connect(mapStateToProps)(Notification);
export {connected as Notification};