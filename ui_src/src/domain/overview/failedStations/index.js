// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { KeyboardArrowRightRounded } from '@material-ui/icons';

import { numberWithCommas, parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import Button from '../../../components/button';
import Filter from '../../../components/filter';
import NoStations from '../../../assets/images/noStations.svg';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { Virtuoso } from 'react-virtuoso';
import HealthyBadge from '../../../components/healthyBadge';

const FailedStations = ({ createStationTrigger }) => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();

    const goToStation = (stationName) => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };
    const Item = React.forwardRef((props, ref) => {
        return <div className="item-wrapper" {...props} ref={ref} />;
    });
    return (
        <div className="overview-wrapper failed-stations-container">
            <p className="overview-components-header">Stations {state?.monitor_data?.stations?.length > 0 && `(${state?.monitor_data?.stations?.length})`}</p>
            <div className="err-stations-list">
                {state?.monitor_data?.stations?.length > 0 ? (
                    <>
                        <div className="coulmns-table">
                            <span style={{ width: '200px' }}>Name</span>
                            <span style={{ width: '200px' }}>Creation date</span>
                            <span style={{ width: '150px' }}>Total messages</span>
                            <span style={{ width: '45px' }}>Status</span>
                            <span style={{ width: '120px' }}></span>
                        </div>
                        <div className="rows-wrapper">
                            <Virtuoso
                                data={state?.monitor_data?.stations}
                                overscan={100}
                                className="testt"
                                components={{ Item }}
                                itemContent={(index, station) => (
                                    <div className="stations-row" key={index} onClick={() => goToStation(station.name)}>
                                        <OverflowTip className="station-details" text={station.name} width={'200px'}>
                                            {station.name}
                                        </OverflowTip>
                                        <OverflowTip className="station-creation" text={parsingDate(station.creation_date)} width={'200px'}>
                                            {parsingDate(station.creation_date)}
                                        </OverflowTip>
                                        <OverflowTip className="station-details total" text={numberWithCommas(station.total_messages)} width={'150px'}>
                                            {numberWithCommas(station.total_messages)}
                                        </OverflowTip>
                                        <span className="station-details status">
                                            <HealthyBadge icon={true} status={station?.has_dls_messages ? 0 : 1} />
                                        </span>
                                        <div className="link-wrapper">
                                            <div className="staion-link">
                                                <span>View Station</span>
                                                <KeyboardArrowRightRounded />
                                            </div>
                                        </div>
                                    </div>
                                )}
                            />
                        </div>
                    </>
                ) : (
                    <div className="empty-stations-container">
                        <img src={NoStations} alt="no stations" onClick={() => createStationTrigger(true)} />
                        <p>No station exist</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder={'Create new station'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => createStationTrigger(true)}
                        />
                    </div>
                )}
            </div>
        </div>
    );
};

export default FailedStations;
