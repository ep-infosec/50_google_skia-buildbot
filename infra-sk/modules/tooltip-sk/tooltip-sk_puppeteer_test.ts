import { expect } from 'chai';
import {
  inBazel, loadCachedTestBed, takeScreenshot, TestBed,
} from '../../../puppeteer-tests/util';

describe('tooltip-sk', () => {
  let testBed: TestBed;
  before(async () => {
    testBed = await loadCachedTestBed();
  });

  beforeEach(async () => {
    // Remove the /dist/ below for //infra-sk elements.
    await testBed.page.goto(inBazel() ? testBed.baseUrl : `${testBed.baseUrl}/tooltip-sk.html`);
    await testBed.page.setViewport({ width: 600, height: 300 });
  });

  it('should render the demo page (smoke test)', async () => {
    expect(await testBed.page.$$('tooltip-sk')).to.have.length(1);
  });

  describe('screenshots', () => {
    it('shows the default view', async () => {
      await takeScreenshot(testBed.page, 'infra-sk', 'tooltip-sk');
    });
  });
});
