/**
 * Copyright 2020 The Magma Authors.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @flow strict-local
 * @format
 */
import type {Dataset} from './CustomMetrics';
import type {network_id} from '@fbcnms/magma-api';

import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';
import LoadingFiller from '@fbcnms/ui/components/LoadingFiller';
import MagmaV1API from '@fbcnms/magma-api/client/WebClient';
import React from 'react';
import moment from 'moment';
import nullthrows from '@fbcnms/util/nullthrows';

import {CustomLineChart, getStep, getStepString} from './CustomMetrics';
import {colors} from '../theme/default';
import {useEffect, useState} from 'react';
import {useEnqueueSnackbar} from '@fbcnms/ui/hooks/useSnackbar';
import {useRouter} from '@fbcnms/ui/hooks';

type Props = {
  startEnd: [moment, moment],
};

type DatasetFetchProps = {
  networkId: network_id,
  start: moment,
  end: moment,
  enqueueSnackbar: (msg: string, cfg: {}) => ?(string | number),
};

async function getEventAlertDataset(props: DatasetFetchProps) {
  const {start, end, networkId} = props;
  const [delta, unit] = getStep(start, end);
  let requestError = '';
  const queries = [];

  let s = start.clone();
  while (end.diff(s) >= 0) {
    const e = s.clone();
    e.add(delta, unit);
    queries.push([s, e]);
    s = e.clone();
  }

  const requests = queries.map(async (query, _) => {
    try {
      const [s, e] = query;
      const response = await MagmaV1API.getEventsByNetworkIdAboutCount({
        networkId: networkId,
        start: s.toISOString(),
        end: e.toISOString(),
      });
      return response;
    } catch (error) {
      requestError = error;
    }
    return null;
  });

  // get events data
  const eventData = await Promise.all(requests)
    .then(allResponses => {
      return allResponses.map((r, index) => {
        const [s] = queries[index];

        if (r === null || r === undefined) {
          return {
            t: s.unix(),
            y: 0,
          };
        }
        return {
          t: s.unix() * 1000,
          y: r,
        };
      });
    })
    .catch(error => {
      requestError = error;
      return [];
    });

  const alertsData = [];
  try {
    const alertPromResp = await MagmaV1API.getNetworksByNetworkIdPrometheusQueryRange(
      {
        networkId: networkId,
        start: start.toISOString(),
        end: end.toISOString(),
        step: getStepString(delta, unit),
        query: 'sum(ALERTS)',
      },
    );
    alertPromResp.data.result.forEach(it =>
      it['values']?.map(i => {
        alertsData.push({
          t: parseInt(i[0]) * 1000,
          y: parseFloat(i[1]),
        });
      }),
    );
  } catch (error) {
    requestError = error;
    return [];
  }

  if (requestError) {
    props.enqueueSnackbar('Error getting event counts', {
      variant: 'error',
    });
  }

  return [
    {
      label: 'Alerts',
      fill: false,
      lineTension: 0.2,
      pointHitRadius: 10,
      pointRadius: 0.1,
      borderWidth: 2,
      backgroundColor: colors.data.flamePea,
      borderColor: colors.data.flamePea,
      hoverBackgroundColor: colors.data.flamePea,
      hoverBorderColor: 'black',
      data: alertsData,
    },
    {
      label: 'Events',
      fill: false,
      backgroundColor: colors.secondary.dodgerBlue,
      borderColor: colors.secondary.dodgerBlue,
      borderWidth: 1,
      hoverBackgroundColor: colors.secondary.dodgerBlue,
      hoverBorderColor: 'black',
      data: eventData,
    },
  ];
}

export default function EventAlertChart(props: Props) {
  const {match} = useRouter();
  const networkId: string = nullthrows(match.params.networkId);
  const [start, end] = props.startEnd;
  const enqueueSnackbar = useEnqueueSnackbar();
  const [isLoading, setIsLoading] = useState(true);

  const [eventDataset, setEventDataset] = useState<Dataset>({
    label: 'Events',
    backgroundColor: colors.secondary.dodgerBlue,
    borderColor: colors.secondary.dodgerBlue,
    borderWidth: 1,
    hoverBackgroundColor: colors.secondary.malibu,
    hoverBorderColor: 'black',
    data: [],
    fill: false,
  });

  const [alertDataset, setAlertDataset] = useState<Dataset>({
    label: 'Alerts',
    backgroundColor: colors.data.flamePea,
    borderColor: colors.data.flamePea,
    borderWidth: 1,
    hoverBackgroundColor: colors.data.flamePea,
    hoverBorderColor: 'black',
    data: [],
    fill: false,
  });

  useEffect(() => {
    // fetch queries
    const fetchAllData = async () => {
      const [eventDataset, alertDataset] = await getEventAlertDataset({
        start,
        end,
        networkId,
        enqueueSnackbar,
      });
      setEventDataset(eventDataset);
      setAlertDataset(alertDataset);
      setIsLoading(false);
    };

    fetchAllData();
  }, [start, end, enqueueSnackbar, networkId]);

  if (isLoading) {
    return <LoadingFiller />;
  }

  return (
    <Card elevation={0}>
      <CardHeader
        subheader={
          <CustomLineChart
            dataset={[eventDataset, alertDataset]}
            yLabel={'count'}
          />
        }
      />
    </Card>
  );
}