import './index';
import { expect } from 'chai';
import { HistogramSk } from './histogram-sk';

import { setUpElementUnderTest } from '../../../infra-sk/modules/test_util';

describe('histogram-sk', () => {
  const newInstance = setUpElementUnderTest<HistogramSk>('histogram-sk');

  let element: HistogramSk;
  beforeEach(() => {
    element = newInstance((el: HistogramSk) => {
      // Place here any code that must run after the element is instantiated but
      // before it is attached to the DOM (e.g. property setter calls,
      // document-level event listeners, etc.).
    });
  });

  describe('some action', () => {
    it('some result', () => {});
    expect(element).to.not.be.null;
  });
});
