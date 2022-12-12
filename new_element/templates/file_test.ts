import './index';
import { expect } from 'chai';
import { {{.ClassName }} } from './{{.ElementName}}';

import { setUpElementUnderTest } from '../../../infra-sk/modules/test_util';

describe('{{.ElementName}}', () => {
  const newInstance = setUpElementUnderTest<{{.ClassName}}>('{{.ElementName}}');

  let element: {{.ClassName}};
  beforeEach(() => {
    element = newInstance((el: {{.ClassName}}) => {
      // Place here any code that must run after the element is instantiated but
      // before it is attached to the DOM (e.g. property setter calls,
      // document-level event listeners, etc.).
    });
  });

  describe('some action', () => {
    it('some result', () => {
      expect(element).to.not.be.null;
    });
  });
});
