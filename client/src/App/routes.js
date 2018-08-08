import {LoginPage} from '../LoginPage';
import {ProductsPage} from '../ProductsPage';
import {RequestsPage} from '../RequestsPage';
import {HomePage} from '../HomePage';

export const publicRoutes = [{
  component: HomePage,
  path: '/'
}, {
  component: LoginPage,
  path: '/login'
}];

export const privateRoutes = [{
  component: ProductsPage,
  path: '/products'
}, {
  component: RequestsPage,
  path: '/requests'
}];