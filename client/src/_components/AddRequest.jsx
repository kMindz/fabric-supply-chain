import React from 'react';
import {connect} from 'react-redux';

import {requestActions} from '../_actions';

class AddRequest extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      request: {
        comment: '',
      },
      submitted: false
    };

    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);

    this.props.setSubmitFn && this.props.setSubmitFn(this.handleSubmit);
  }

  handleChange(event) {
    const {name, value} = event.target;
    const {request} = this.state;
    this.setState({
      request: {
        ...request,
        [name]: value
      }
    });
  }

  handleSubmit(event) {
    event.preventDefault();

    this.setState({submitted: true});
    const {request} = this.state;
    if (request.comment) {
      this.props.dispatch(requestActions.add(this.props.modal.object, request.comment));
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