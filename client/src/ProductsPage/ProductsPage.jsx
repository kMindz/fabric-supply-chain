import React from 'react';
import ReactTable from 'react-table';
import {connect} from 'react-redux';

import {productActions, modalActions} from '../_actions';
import {AddProduct, AddRequest, HistoryTable, Modal} from '../_components';
import {productStates, orgConstants} from '../_constants';

const modalIds = {
  addProduct: 'addProduct',
  editProduct: 'editProduct',
  history: 'history',
  addRequest: 'addRequest'
};

const productHistoryColumns = [{
  Header: 'Owner',
  id: 'value.owner',
  accessor: rec => orgConstants[rec.value.owner]
}, {
  Header: 'Description',
  accessor: 'value.desc'
}, {
  id: 'state',
  Header: 'State',
  accessor: rec => productStates[rec.value.state],
  filterMethod: (filter, row) => {
    if (filter.value === "all") {
      return true;
    }
    return productStates[+filter.value] === row.state;
  },
  Filter: ({filter, onChange}) =>
    <select
      onChange={event => onChange(event.target.value)}
      style={{width: "100%"}}
      value={filter ? filter.value : "all"}
    >
      <option value="all">All</option>
      {Object.entries(productStates).map(e => {
        let [k, v] = e;
        return (<option value={k}>{v}</option>);
      })}
    </select>
}, {
  id: 'lastUpdated',
  Header: 'Updated',
  accessor: rec => new Date(rec.value.lastUpdated).toLocaleString(),
  filterMethod: (filter, row) => {
    return row.lastUpdated && row.lastUpdated.indexOf(filter.value) > -1;
  }
}];

class ProductsPage extends React.Component {
  constructor() {
    super();

    this.loadHistory = this.loadHistory.bind(this);
    this.refreshData = this.refreshData.bind(this);
  }

  componentDidUpdate(prevProps) {
    const {modals, products, dispatch, requests} = this.props;
    const modalProps = modals[modalIds.history];
    if (modalProps.show && products.history) {
      const {[modalProps.object.key.name]: data} = products.history;
      const {[modalProps.object.key.name]: prevData} = prevProps.products.history || {};
      if (prevData !== data) {
        dispatch(modalActions.setData(modalIds.history, data));
      }
    }

    //1st load
    if (!prevProps.products.hasOwnProperty('adding')) {
      this.refreshData();
    }

    if (products.adding === false) {
      this.refreshData();
      dispatch(modalActions.hide(modalIds.addProduct));
      dispatch(modalActions.hide(modalIds.editProduct));
    }

    if (requests && requests.adding === false && prevProps.requests.adding !== requests.adding) {
      dispatch(modalActions.hide(modalIds.addRequest));
    }
  }

  refreshData() {
    this.props.dispatch(productActions.getAll());
  }

  loadHistory(product) {
    this.props.dispatch(productActions.history(product));
  }

  render() {
    const {products, user} = this.props;

    if (!products) {
      return null;
    }

    const columns = [{
      Header: 'Name',
      accessor: 'key.name'
    }, ...productHistoryColumns, {
      id: 'actions',
      Header: 'Actions',
      accessor: 'id',
      sortable: false,
      filterable: false,
      Cell: row => {
        return (
          <div>
            <button className="btn btn-sm btn-secondary" title="History"
              onClick={Modal.open.bind(this, modalIds.history, row.original)}>
              <i className="fas fa-fw fa-clipboard-list"/>
            </button>
            {row.original.value.owner === user.org &&
            (<button className="btn btn-sm btn-primary" title="Edit"
                     onClick={Modal.open.bind(this, modalIds.editProduct, row.original)}>
              <i className="fas fa-fw fa-pen"/>
            </button>)
            }
            {row.original.value.owner !== user.org && (
              <button className="btn btn-sm btn-primary" title="Request"
                      onClick={Modal.open.bind(this, modalIds.addRequest, row.original)}>
                <i className="fas fa-fw fa-plus"/>
              </button>)}
          </div>
        )
      }
    }];

    return (
      <div>
        <button className="btn btn-primary" onClick={Modal.open.bind(this, modalIds.addProduct)}>
          Add new product
        </button>
        <Modal modalId="history" title="History" large={true} footer={false}>
          <HistoryTable columns={productHistoryColumns}
            loadData={this.loadHistory}
            defaultSorted={[
              {
                id: "lastUpdated",
                desc: false
              }
            ]}
          />
        </Modal>
        <Modal modalId="addProduct" title="Add new product">
          <AddProduct/>
        </Modal>
        <Modal modalId="editProduct" title="Edit product">
          <AddProduct/>
        </Modal>
        <Modal modalId="addRequest" title="Request ownership">
          <AddRequest/>
        </Modal>
        <hr/>
        <h3>All products <button className="btn" onClick={this.refreshData}><i className="fas fa-sync"/></button></h3>
        <ReactTable
          columns={columns}
          data={products.items}
          className="-striped -highlight"
          defaultPageSize={10}
          filterable={true}
          defaultSorted={[
            {
              id: "lastUpdated",
              desc: true
            }
          ]}
          // freezeWhenExpanded={true}
          // collapseOnDataChange={false}
          // SubComponent={row => (
          //   <div>{products.history[row.original.key.name].value.owner}</div>
          // )}
          // getTrProps={(state, rowInfo, column, instance) => {
          //   return {
          //     onClick: e => {
          //       this.props.dispatch(productActions.history(rowInfo.original));
          //     }
          //   }
          // }}
        />
      </div>
    );
  }
}

function mapStateToProps(state) {
  const {products, authentication, modals, requests} = state;
  const {user} = authentication;
  return {
    user,
    products,
    modals,
    requests
  };
}

const connected = connect(mapStateToProps)(ProductsPage);
export {connected as ProductsPage};