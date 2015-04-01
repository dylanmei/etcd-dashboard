FROM progrium/busybox

ADD dashboard/dist /dashboard/dist
ADD bin/etcd-dashboard /etcd-dashboard

ENV ETCD_ADDR localhost:4001
EXPOSE 8080 
ENTRYPOINT [ "/etcd-dashboard" ]
CMD [ "-etcd-addr", "$ETCD_ADDR" ]
