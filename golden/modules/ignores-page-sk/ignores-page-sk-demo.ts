import './index';
import '../gold-scaffold-sk';

import fetchMock from 'fetch-mock';
import { delay } from '../demo_util';
import { ignoreRules_10, fakeNow } from './test_data';
import { manyParams } from '../shared_demo_data';
import { testOnlySetSettings } from '../settings';
import { exampleStatusData } from '../last-commit-sk/demo_data';
import { GoldScaffoldSk } from '../gold-scaffold-sk/gold-scaffold-sk';

Date.now = () => fakeNow;
testOnlySetSettings({
  title: 'Skia Public',
  baseRepoURL: 'https://skia.googlesource.com/skia.git',
});

fetchMock.get('/json/v2/paramset', delay(manyParams, 100));
fetchMock.get('/json/v2/ignores', delay(ignoreRules_10, 300));
fetchMock.post('glob:/json/v1/ignores/del/*', delay({}, 600));
fetchMock.post('glob:/json/v1/ignores/add/', delay({}, 600));
fetchMock.post('glob:/json/v1/ignores/save/*', delay({}, 600));
fetchMock.get('/json/v2/trstatus', JSON.stringify(exampleStatusData));

// By adding these elements after all the fetches are mocked out, they should load ok.
const newScaf = new GoldScaffoldSk();
newScaf.testingOffline = true;
// Make it the first element in body.
document.body.insertBefore(newScaf, document.body.childNodes[0]);
const page = document.createElement('ignores-page-sk');
page.setAttribute('page_size', '10');
newScaf.appendChild(page);
