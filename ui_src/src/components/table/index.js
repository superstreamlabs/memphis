// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { Table as CustomTable } from 'antd';
import React, { useEffect, useState } from 'react';

const Table = ({ columns, data, title, tableRowClassname, className, onSelectRow }) => {
    const [windowHeight, setWindowHeight] = useState(window.innerHeight);
    const [numRows, setNumRows] = useState(10);

    useEffect(() => {
        const handleResize = () => {
            setWindowHeight(window.innerHeight);
        };

        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    useEffect(() => {
        const calculateNumRows = () => {
            const rowHeight = 88; // Adjust this value based on your table's row height
            const maxNumRows = Math.floor(windowHeight / rowHeight) - 1; // Subtract 1 to account for the table header
            setNumRows(maxNumRows);
        };

        calculateNumRows();
    }, [windowHeight]);

    const itemRender = (_, type, originalElement) => {
        if (type === 'prev') {
            return <a className="a-link ">Previous</a>;
        }
        if (type === 'next') {
            return <a className="a-link ">Next</a>;
        }
        return originalElement;
    };

    const fieldProps = {
        columns,
        dataSource: data,
        title,
        rowClassName: tableRowClassname,
        className
    };

    return (
        <CustomTable
            {...fieldProps}
            pagination={{ pageSize: numRows, itemRender: itemRender, hideOnSinglePage: true, responsive: false }}
            rowKey={(record) => record.id}
            onRow={(record, rowIndex) => {
                return {
                    onClick: (event) => {
                        onSelectRow(record, rowIndex);
                    }
                };
            }}
        />
    );
};

export default Table;
