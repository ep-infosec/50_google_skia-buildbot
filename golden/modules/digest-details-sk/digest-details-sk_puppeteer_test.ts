import { expect } from 'chai';
import { loadCachedTestBed, takeScreenshot, TestBed } from '../../../puppeteer-tests/util';

describe('digest-details-sk', () => {
  let testBed: TestBed;
  before(async () => {
    testBed = await loadCachedTestBed();
  });

  beforeEach(async () => {
    await testBed.page.goto(testBed.baseUrl, { waitUntil: 'networkidle0' });
  });

  it('should render the demo page', async () => {
    // Smoke test.
    expect(await testBed.page.$$('digest-details-sk')).to.have.length(9);
  });

  describe('screenshots', () => {
    it('has the left and right image', async () => {
      const digestDetailsSk = await testBed.page.$('#normal');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk');
    });

    it('was given data with only a negative image to compare against', async () => {
      const digestDetailsSk = await testBed.page.$('#negative_only');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_negative-only');
    });

    it('was given no other images to compare against', async () => {
      const digestDetailsSk = await testBed.page.$('#no_refs');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_no-refs');
    });

    it('is computing the closest positive and negative', async () => {
      const digestDetailsSk = await testBed.page.$('#no_refs_yet');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_computing-refs');
    });

    it('was given a changelist id', async () => {
      const digestDetailsSk = await testBed.page.$('#changelist_id');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_changelist-id');
    });

    it('had the right side overridden', async () => {
      const digestDetailsSk = await testBed.page.$('#right_overridden');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_right-overridden');
    });

    it('had no trace data sent by the backend', async () => {
      const digestDetailsSk = await testBed.page.$('#no_traces');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_no-traces');
    });

    it('had no params sent by the backend', async () => {
      const digestDetailsSk = await testBed.page.$('#no_params');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_no-params');
    });

    it('shows full size images', async () => {
      const digestDetailsSk = await testBed.page.$('#full_size_images');
      await takeScreenshot(digestDetailsSk!, 'gold', 'digest-details-sk_full-size-images');
    });
  });
});
