import React from 'react';
import ReactTable from 'react-table';

export class HistoryTable extends React.Component {
  componentDidMount() {
    //TODO
    this.props.loadData(this.props.modal.object);
  }

  render() {
    const {children, ...rest} = this.props;

    return (
      <ReactTable
        className="-striped -highlight"
        defaultPageSize={10}
        filterable={true}
        {...rest}
      />
    );
  }
}