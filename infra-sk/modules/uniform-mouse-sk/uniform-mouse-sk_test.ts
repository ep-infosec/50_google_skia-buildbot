import './index';
import { assert } from 'chai';
import { UniformMouseSk } from './uniform-mouse-sk';

import { setUpElementUnderTest } from '../test_util';

describe('uniform-mouse-sk', () => {
  const newInstance = setUpElementUnderTest<UniformMouseSk>('uniform-mouse-sk');

  let element: UniformMouseSk;
  beforeEach(() => {
    element = newInstance();
  });

  describe('uniform-mouse-sk', () => {
    it('reports uniform values correctly', () => {
      const uniforms: number[] = new Array(4);
      element.applyUniformValues(uniforms);
      assert.deepEqual(uniforms, [0, 0, -1, -1]);
    });

    it('throws on an invalid uniform', () => {
      assert.throws(() => {
        element.uniform = {
          name: '',
          rows: 2,
          columns: 2,
          slot: 1,
        };
      });
    });

    it('does not need raf updates', () => {
      assert.isFalse(element.needsRAF());
    });
  });
});
