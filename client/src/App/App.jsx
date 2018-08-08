import React from 'react';
import {Route, Switch, HashRouter} from 'react-router-dom';
import {connect} from 'react-redux';

import {history} from '../_helpers';
import {alertActions, configActions} from '../_actions';
import {Footer, Header, Main} from '../Layout';
import {publicRoutes, privateRoutes} from './routes';
import {ScaleLoader} from 'react-spinners';

class App extends React.Component {
  constructor(props) {
    super(props);

    const {dispatch} = this.props;
    history.listen((location, action) => {
      // clear alert on location change
      dispatch(alertActions.clear());
    });
    dispatch(configActions.get());
  }

  render() {
    if (!this.props.config) {
      return null;
    }

    return (
      <HashRouter>
        <Switch>
          {publicRoutes.concat(privateRoutes).map((route, key) => {
            const {component, path} = route;
            return (
              <Route
                exact
                path={path}
                key={key}
                render={(route) =>
                  <div>
                    <Header route={route}/>
                    <div className={'overlay d-flex justify-content-center align-items-center ' + (!!this.props.loading ? 'visible' : 'invisible')}>
                      <ScaleLoader loading={!!this.props.loading} />
                    </div>
                    <Main component={component} route={route} config={this.props.config}/>
                    <Footer/>
                  </div>}
              />
            )
          })}
        </Switch>
      </HashRouter>
    );
  }
}

function mapStateToProps(state) {
  const {config, loading} = state;
  return {
    config,
    loading
  };
}

const connectedApp = connect(mapStateToProps)(App);
export {connectedApp as App};