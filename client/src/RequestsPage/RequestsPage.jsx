import React from 'react';
import ReactTable from 'react-table';
import {connect} from 'react-redux';

import {requestActions, modalActions} from '../_actions';
import {HistoryTable, Modal} from '../_components';
import {orgConstants} from '../_constants';

const requestHistoryColumns = [{
  Header: 'Request Sender',
  id: 'key.requestSender',
  accessor: rec => orgConstants[rec.key.requestSender]
}, {
  Header: 'Request Receiver',
  id: 'key.requestReceiver',
  accessor: rec => orgConstants[rec.key.requestReceiver]
}, {
  Header: 'Status',
  accessor: 'value.status',
  Cell: row => {
    const classMap = {
      'Initiated': '',
      'Accepted': 'success',
      'Rejected': 'danger',
      'Cancelled': 'warning'
    };
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
      {['Initiated', 'Accepted', 'Rejected', 'Cancelled'].map(v => {
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
  }
}];

class RequestsPage extends React.Component {
  constructor() {
    super();

    this.handleOpenModal = this.handleOpenModal.bind(this);
    this.loadHistory = this.loadHistory.bind(this);
    this.refreshData = this.refreshData.bind(this);
  }

  componentDidMount() {
    this.refreshData();

    this.acceptRequest = this.acceptRequest.bind(this);
    this.rejectRequest = this.rejectRequest.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    const {requests, dispatch, modals} = nextProps;
    if (requests.adding === false) {
      this.refreshData();
    }

    if (modals) {
      const {historyReq} = modals;
      if (historyReq && historyReq.show && nextProps.requests.history) {
        this.historyData = nextProps.requests.history[historyReq.object.key.productKey];
      }
    }
  }

  refreshData() {
    this.props.dispatch(requestActions.getAll());
  }

  handleOpenModal(modalId, product) {
    this.props.dispatch(modalActions.show(modalId, product));
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
    }, {
      Header: 'Product owner',
      id: 'key.requestReceiver',
      accessor: rec => orgConstants[rec.key.requestReceiver]
    }, {
      Header: 'Requester',
      id: 'key.requestSender',
      accessor: rec => orgConstants[rec.key.requestSender]
    }, {
      Header: 'Status',
      accessor: 'value.status',
      Cell: row => {
        const classMap = {
          'Initiated': '',
          'Accepted': 'success',
          'Rejected': 'danger',
          'Cancelled': 'warning'
        };
        return (<div className={'bg-' + classMap[row.value]}>{row.value}</div>);
      }
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
    }, {
      id: 'actions',
      Header: 'Actions',
      accessor: 'key.productKey',
      Cell: row => {
        const record = row.original;
        return (
          <div>
            <button className="btn btn-sm btn-primary" title="History"
                    onClick={this.handleOpenModal.bind(this, 'historyReq', row.original)}>
              <i className="fas fa-fw fa-history"/>
            </button>
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
                        data={this.historyData}
                        defaultSorted={[
                          {
                            id: "timestamp",
                            desc: true
                          }
                        ]}/>
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