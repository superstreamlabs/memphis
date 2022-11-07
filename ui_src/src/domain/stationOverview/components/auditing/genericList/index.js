// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useContext, useEffect, useState } from 'react';

import OverflowTip from '../../../../../components/tooltip/overflowtip';
import Reducer from '../../../hooks/reducer';
import { StationStoreContext } from '../../..';
import { parsingDate } from '../../../../../services/valueConvertor';

const GenericList = (props) => {
    const [stationState] = useContext(StationStoreContext);

    const { columns, tab } = props;
    const [rowsData, setRowsData] = useState([]);

    useEffect(() => {
        if (tab === 0) {
            setRowsData(stationState?.stationSocketData?.audit_logs);
        }
    }, [stationState]);

    return (
        <div className="generic-list-wrapper">
            <div className="list">
                <div className="coulmns-table">
                    {columns?.map((column, index) => {
                        return (
                            <span key={index} style={{ width: column.width }}>
                                {column.title}
                            </span>
                        );
                    })}
                </div>
                <div className="rows-wrapper">
                    {rowsData?.map((row, index) => {
                        return (
                            <div className="pubSub-row" key={index}>
                                <OverflowTip text={row?.message || row?.produced_by} width={'300px'}>
                                    {row?.message || row?.produced_by}
                                </OverflowTip>
                                <OverflowTip text={row?.created_by_user || row?.consumer} width={'200px'}>
                                    {row?.created_by_user || row?.consumer}
                                </OverflowTip>
                                <OverflowTip text={parsingDate(row?.creation_date)} width={'200px'}>
                                    {parsingDate(row?.creation_date)}
                                </OverflowTip>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

export default GenericList;
