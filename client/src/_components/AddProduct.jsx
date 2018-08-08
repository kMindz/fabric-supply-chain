import React from 'react';
import {connect} from 'react-redux';

import {productActions} from '../_actions';
import {productStates} from '../_constants';

class AddProduct extends React.Component {
  constructor(props) {
    super(props);

    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);

    this.props.setSubmitFn && this.props.setSubmitFn(this.handleSubmit);

    this.state = {
      product: {
        name: '',
        desc: '',
        state: 1,
        created: false
      },
      submitted: false
    };
    this._fillProduct();
  }

  _fillProduct() {
    if(this.props.initData && this.props.initData.key) {
      this.state.product.name = this.props.initData.key.name;
      this.state.product.desc = this.props.initData.value.desc;
      this.state.product.state = this.props.initData.value.state;
      this.state.product.created = true;
    }
  }

  handleChange(event) {
    const {name, value} = event.target;
    const {product} = this.state;
    this.setState({
      product: {
        ...product,
        [name]: value
      },
      submitted: false
    });
  }

  handleSubmit(event) {
    event.preventDefault();

    this.setState({submitted: true});
    const {product} = this.state;
    if (product.name) {
      this.props.dispatch(productActions[product.created ? 'edit' : 'add'](product));
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
          <div className="text-danger">Name is required</div>
          }
        </div>
        <div>
          <label htmlFor="desc">Description</label>
          <textarea type="text" className="form-control" name="desc" value={product.desc}
                    onChange={this.handleChange}/>
        </div>
        {product.created && <div>
          <label htmlFor="state">State</label>
          <select className="form-control" name="state" value={product.state}
                  onChange={this.handleChange}>
            {Object.entries(productStates).map(e => {
              let [k, v] = e;
              return (<option value={k}>{v}</option>);
            })}
          </select>
        </div>}
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

const connected = connect(mapStateToProps)(AddProduct);
export {connected as AddProduct};