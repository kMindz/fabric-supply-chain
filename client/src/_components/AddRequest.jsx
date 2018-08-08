import React from 'react';
import {connect} from 'react-redux';

import {requestActions} from '../_actions';

class AddRequest extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      request: {
        comment: '',
        created: false
      },
      submitted: false
    };
    this._fillRequest();

    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);

    this.props.setSubmitFn && this.props.setSubmitFn(this.handleSubmit);
  }

  _fillRequest() {
    if(this.props.initData && this.props.initData.key.productKey) {
      this.state.request.comment = this.props.initData.value.message;
      this.state.request.created = true;
    }
  }

  handleChange(event) {
    const {name, value} = event.target;
    const {request} = this.state;
    this.setState({
      request: {
        ...request,
        [name]: value
      },
      submitted: false
    });
  }

  handleSubmit(event) {
    event.preventDefault();

    this.setState({submitted: true});
    const {request} = this.state;
    if (request.comment) {
      this.props.dispatch(requestActions[request.created ? 'edit' : 'add'](this.props.initData, request.comment));
    }
  }


  render() {
    const {request, submitted} = this.state;
    return (
      <form name="form" onSubmit={this.handleSubmit}>
        <div className={'form-group'}>
          <label htmlFor="comment">Comment</label>
          <textarea className={"form-control" + (submitted && !request.comment ? ' is-invalid' : '')}
                    name="comment" value={request.comment}
                    onChange={this.handleChange}/>
          {submitted && !request.comment &&
          <div className="text-danger form-text">Comment is required</div>
          }
        </div>
      </form>
    );
  }
}

function mapStateToProps(state) {
  const {request} = state;

  return {
    request
  }
}

const connected = connect(mapStateToProps)(AddRequest);
export {connected as AddRequest};