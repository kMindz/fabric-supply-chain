import React from 'react';
import {connect} from 'react-redux';

import {productActions} from '../_actions';

class AddProduct extends React.Component {
  constructor(props) {
    super(props);

    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);

    this.props.setSubmitFn && this.props.setSubmitFn(this.handleSubmit);

    this.state = {
      product: {
        name: '',
        desc: ''
      },
      submitted: false
    };
  }

  handleChange(event) {
    const {name, value} = event.target;
    const {product} = this.state;
    this.setState({
      product: {
        ...product,
        [name]: value
      }
    });
  }

  handleSubmit(event) {
    event.preventDefault();

    this.setState({submitted: true});
    const {product} = this.state;
    if (product.name) {
      this.props.dispatch(productActions.add(product));
    }
  }

  render() {
    const {product, submitted} = this.state;
    return (
      <form name="form" onSubmit={this.handleSubmit}>
        <div className={'form-group'}>
          <label htmlFor="name">Name</label>
          <input type="text" className={"form-control" + (submitted && !product.name ? ' is-invalid' : '')}
                 name="name" value={product.name}
                 onChange={this.handleChange}/>
          {submitted && !product.name &&
          <div className="text-danger form-text">Name is required</div>
          }
        </div>
        <div>
          <label htmlFor="desc">Description</label>
          <textarea type="text" className="form-control" name="desc" value={product.desc}
                    onChange={this.handleChange}/>
        </div>
      </form>
    );
  }
}

function mapStateToProps(state) {
  const {product} = state;
  return {
    product
  }
}

const connectedAddProduct = connect(mapStateToProps)(AddProduct);
export {connectedAddProduct as AddProduct};