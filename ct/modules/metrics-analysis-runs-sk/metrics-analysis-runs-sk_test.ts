import './index';

import sinon from 'sinon';
import { expect } from 'chai';
import { $, $$ } from 'common-sk/modules/dom';
import fetchMock from 'fetch-mock';
import { MetricsAnalysisRunsSk } from './metrics-analysis-runs-sk';

import {
  tasksResult0, tasksResult1,
} from './test_data';
import {
  eventPromise,
  setUpElementUnderTest,
} from '../../../infra-sk/modules/test_util';

describe('metrics-analysis-runs-sk', () => {
  const newInstance = setUpElementUnderTest<MetricsAnalysisRunsSk>('metrics-analysis-runs-sk');
  fetchMock.config.overwriteRoutes = false;

  let analysisRuns: HTMLElement;
  beforeEach(async () => {
    await expectReload(() => analysisRuns = newInstance(), null);
  });

  afterEach(() => {
    //  Check all mock fetches called at least once and reset.
    expect(fetchMock.done()).to.be.true;
    fetchMock.reset();
    sinon.restore();
  });

  // Expect 'trigger' to cause a reload, and execute it.
  // Optionally pass desired result from server.
  const expectReload = async (trigger: any, result: any) => {
    result = result || tasksResult0;
    const event = eventPromise('end-task');
    fetchMock.postOnce('begin:/_/get_metrics_analysis_tasks', result);
    trigger();
    await event;
  };

  it('shows table entries', async () => {
    expect($('table.runssummary>tbody>tr', analysisRuns)).to.have.length(11);
    expect(fetchMock.lastUrl()).to.contain('offset=0');
    expect(fetchMock.lastUrl()).to.contain('size=10');
    expect(fetchMock.lastUrl()).to.not.contain('filter_by_logged_in_user=true');
  });

  it('filters by user', async () => {
    expect(fetchMock.lastUrl()).to.not.contain('filter_by_logged_in_user=true');
    await expectReload(() => ($$('#userFilter', analysisRuns)! as HTMLElement).click(), null);
    expect(fetchMock.lastUrl()).to.contain('filter_by_logged_in_user=true');
  });

  it('navigates with pages', async () => {
    expect(fetchMock.lastUrl()).to.contain('offset=0');
    const result = tasksResult1;
    result.pagination!.offset = 10;
    // 'Next page' button.
    await expectReload(
      () => ($('pagination-sk button.action', analysisRuns)[2] as HTMLElement).click(), result,
    );
    expect(fetchMock.lastUrl()).to.contain('offset=10');
    expect($('table.runssummary>tbody>tr', analysisRuns)).to.have.length(5);
  });

  it('deletes tasks', async () => {
    sinon.stub(window, 'confirm').returns(true);
    sinon.stub(window, 'alert');
    fetchMock.post('begin:/_/delete_metrics_analysis_task', 200);
    fetchMock.postOnce('begin:/_/get_metrics_analysis_tasks', tasksResult0);
    ($$('delete-icon-sk', analysisRuns) as HTMLElement).click();
    expect(fetchMock.lastOptions('begin:/_/delete')!.body).to.contain('"id":429');
  });

  it('reschedules tasks', async () => {
    sinon.stub(window, 'confirm').returns(true);
    sinon.stub(window, 'alert');
    fetchMock.post('begin:/_/redo_metrics_analysis_task', 200);
    fetchMock.postOnce('begin:/_/get_metrics_analysis_tasks', tasksResult0);
    ($$('redo-icon-sk', analysisRuns) as HTMLElement).click();
    expect(fetchMock.lastOptions('begin:/_/redo')!.body).to.contain('"id":429');
  });

  it('shows detail dialogs', async () => {
    expect($$('.dialog-background', analysisRuns)!.classList.value).to.include('hidden');
    ($$('.details', analysisRuns) as HTMLElement).click();
    expect($$('.dialog-background', analysisRuns)!.classList.value).to.not.include('hidden');
  });
});
