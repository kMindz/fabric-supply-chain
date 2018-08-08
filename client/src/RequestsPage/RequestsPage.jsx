import React from 'react';
import ReactTable from 'react-table';
import {connect} from 'react-redux';

import {requestActions, modalActions} from '../_actions';
import {HistoryTable, AddRequest, Modal} from '../_components';
import {orgConstants} from '../_constants';

const classMap = {
  'Initiated': '',
  'Accepted': 'success',
  'Rejected': 'danger',
  'Cancelled': 'warning'
};

const modalIds = {
  editRequest: 'editRequest',
  history: 'historyReq'
};

const requestHistoryColumns = [{
  Header: 'Request Sender',
  id: 'key.requestSender',
  accessor: rec => orgConstants[rec.key.requestSender]
}, {
  Header: 'Product owner',
  id: 'key.requestReceiver',
  accessor: rec => orgConstants[rec.key.requestReceiver]
}, {
  Header: 'Status',
  accessor: 'value.status',
  Cell: row => {
    return (<div className={'bg-' + classMap[row.value]}>{row.value}</div>);
  },
  filterMethod: (filter, row) => {
    if (filter.value === "all") {
      return true;
    }
    return filter.value === row['value.status'];
  },
  Filter: ({filter, onChange}) =>
    <select
      onChange={event => onChange(event.target.value)}
      style={{width: "100%"}}
      value={filter ? filter.value : "all"}
    >
      <option value="all">All</option>
      {Object.keys(classMap).map(v => {
        return (<option value={v}>{v}</option>);
      })}
    </select>
}, {
  Header: 'Message',
  accessor: 'value.message'
}, {
  id: 'timestamp',
  Header: 'Updated',
  accessor: rec => new Date(rec.value.timestamp * 1000).toLocaleString(),
  filterMethod: (filter, row) => {
    return row.timestamp && row.timestamp.indexOf(filter.value) > -1;
  },
  sortMethod: (a, b) => {
    return a && b && new Date(a).getTime() > new Date(b).getTime() ? 1 : -1;
  }
}];

class RequestsPage extends React.Component {
  constructor() {
    super();

    this.loadHistory = this.loadHistory.bind(this);
    this.refreshData = this.refreshData.bind(this);
    this.acceptRequest = this.acceptRequest.bind(this);
    this.rejectRequest = this.rejectRequest.bind(this);
  }

  componentDidUpdate(prevProps) {
    const {requests, modals, dispatch} = this.props;
    if (modals) {
      const modalProps = modals[modalIds.history];
      if (modalProps && modalProps.show && requests.history) {
        const {[modalProps.object.key.productKey]: data} = requests.history;
        const {[modalProps.object.key.productKey]: prevData} = prevProps.requests.history || {};
        if (prevData !== data) {
          dispatch(modalActions.setData(modalIds.history, data));
        }
      }
    }



    if (!requests.hasOwnProperty('adding')) {
      this.refreshData();
    }

    if (this.props.requests.adding === false) {
      this.refreshData();
      this.props.dispatch(modalActions.hide(modalIds.history));
      this.props.dispatch(modalActions.hide(modalIds.editRequest));
    }
  }

  refreshData() {
    this.props.dispatch(requestActions.getAll());
  }

  loadHistory(request) {
    this.props.dispatch(requestActions.history(request));
  }

  render() {
    const {requests, user} = this.props;

    if(!requests) {
      return null;
    }

    const columns = [{
      Header: 'Name',
      accessor: 'key.productKey'
    }, ...requestHistoryColumns, {
      id: 'actions',
      Header: 'Actions',
      accessor: 'key.productKey',
      Cell: row => {
        const record = row.original;
        return (
          <div>
            <button className="btn btn-sm btn-secondary" title="History"
                    onClick={Modal.open.bind(this, modalIds.history, row.original)}>
              <i className="fas fa-fw fa-clipboard-list"/>
            </button>
            {record.value.status === 'Initiated' && record.key.requestSender === user.org &&
            (<button className="btn btn-sm btn-primary" title="Edit"
                     onClick={Modal.open.bind(this, modalIds.editRequest, row.original)}>
              <i className="fas fa-fw fa-pen"/>
            </button>)
            }
            {record.value.status === 'Initiated' && record.key.requestSender === user.org &&
              (<button className="btn btn-sm btn-warning" title="Cancel"
                       onClick={()=>{this.rejectRequest(row.original)}}>
                <i className="fas fa-fw fa-times"/>
              </button>)
            }
            {(record.value.status === 'Initiated' && record.key.requestReceiver === user.org) &&
              (<button className="btn btn-sm btn-success" title="Accept"
                       onClick={()=>{this.acceptRequest(row.original)}}>
                <i className="fas fa-fw fa-check"/>
              </button>)
            }
            {(record.value.status === 'Initiated' && record.key.requestReceiver === user.org) &&
              (<button className="btn btn-sm btn-danger"  title="Reject"
                       onClick={()=>{this.rejectRequest(row.original)}}>
                  <i className="fas fa-fw fa-times"/>
              </button>)
            }
          </div>
        )
      }
    }];

    return (
      <div>
        <h3>All requests <button className="btn" onClick={this.refreshData}><i className="fas fa-sync"/></button></h3>
        <Modal modalId="historyReq" title="History" large={true} footer={false}>
          <HistoryTable columns={requestHistoryColumns}
                        loadData={this.loadHistory}
                        defaultSorted={[
                          {
                            id: "timestamp",
                            desc: false
                          }
                        ]}/>
        </Modal>
        <Modal modalId="editRequest" title="Edit Request">
          <AddRequest/>
        </Modal>
        {requests.items &&
        <ReactTable
          columns={columns}
          data={requests.items}
          className="-striped -highlight"
          defaultPageSize={10}
          filterable={true}
          defaultSorted={[
            {
              id: "timestamp",
              desc: true
            }
          ]}/>
        }
      </div>
    );
  }

  acceptRequest(record) {
    this.props.dispatch(requestActions.accept(record));
  }

  rejectRequest(record) {
    this.props.dispatch(requestActions.reject(record));
  }
}

function mapStateToProps(state) {
  const {requests, authentication, modals} = state;
  const {user} = authentication;
  return {
    requests,
    user,
    modals
  };
}

const connected = connect(mapStateToProps)(RequestsPage);
export {connected as RequestsPage};